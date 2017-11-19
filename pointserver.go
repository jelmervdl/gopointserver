package main

import (
	"fmt"
	"strings"
	"errors"
	"strconv"
	"os"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"github.com/MadAppGang/kdbush"
	"github.com/paulmach/go.geojson"
)


type Record struct {
	Feature *geojson.Feature
}


type DataSet struct {
	FeatureCollection *geojson.FeatureCollection
	Index *kdbush.KDBush
}


func (p Record) Coordinates() (float64, float64) {
	if !p.Feature.Geometry.IsPoint() {
		panic("Only Point features are supported")
	}
	
	return p.Feature.Geometry.Point[0], p.Feature.Geometry.Point[1]
}


type BoundingBox struct {
	MinX, MinY, MaxX, MaxY float64
}


var (
	data *DataSet
)


func UnmarshalBoundingBox(str string) (BoundingBox, error) {
	components := strings.Split(str, ",")
	if len(components) != 4 {
		return BoundingBox{}, errors.New("bbox string is not 4 components long")
	}

	minX, err := strconv.ParseFloat(components[0], 64)
	if err != nil {
		return BoundingBox{}, fmt.Errorf("Could not decode first component:", err)
	}

	minY, err := strconv.ParseFloat(components[1], 64)
	if err != nil {
		return BoundingBox{}, fmt.Errorf("Could not decode first component:", err)
	}

	maxX, err := strconv.ParseFloat(components[2], 64)
	if err != nil {
		return BoundingBox{}, fmt.Errorf("Could not decode first component:", err)
	}

	maxY, err := strconv.ParseFloat(components[3], 64)
	if err != nil {
		return BoundingBox{}, fmt.Errorf("Could not decode first component:", err)
	}

	return BoundingBox{minX, minY, maxX, maxY}, nil
}


func UnmarshalPoint(str string) (kdbush.Point, error) {
	components := strings.Split(str, ",")

	if len(components) != 2 {
		return nil, errors.New("point string is not 2 components long")
	}

	x, err := strconv.ParseFloat(components[0], 64); 
	if err != nil {
		return nil, fmt.Errorf("Could not decode first component:", err)
	}

	y, err := strconv.ParseFloat(components[1], 64);
	if err != nil {
		return nil, fmt.Errorf("Could not decode second component:", err)
	}

	return &kdbush.SimplePoint{x, y}, nil
}


func makeJSONHandler(fn func(*http.Request) ([]byte, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := fn(r)

		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Error: %s", err)
		} else {
			w.Header().Add("Content-Type", "application/json; charset=utf8")
			w.Header().Add("Content-Length", strconv.Itoa(len(bytes)))
			w.WriteHeader(200)
			w.Write(bytes)	
		}
	}
}


func makeFeatureCollectionHandler(fn func(*DataSet, *http.Request) ([]int, error)) http.HandlerFunc {
	return makeJSONHandler(func(r *http.Request) ([]byte, error) {
		ds := data

		results, err := fn(ds, r)
		if err != nil {
			return nil, err
		}

		rfc := geojson.NewFeatureCollection()

		for _, i := range results {
			rfc.AddFeature(ds.GetFeature(i))
		}

		bytes, err := rfc.MarshalJSON()
		if err != nil {
			return nil, err
		}

		return bytes, nil
	})
}


func NewDataSet(paths []string) (*DataSet, error) {
	ds := new(DataSet)

	ds.FeatureCollection = geojson.NewFeatureCollection()

	/* Try to load the features from each file */
	for _, path := range paths {
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			dat, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			ffc, err := geojson.UnmarshalFeatureCollection(dat)
			if err != nil {
				return err
			}

			ds.AddFeatures(ffc.Features)

			return nil
		})
	}

	/* Create the index */
	points := make([]kdbush.Point, len(ds.FeatureCollection.Features))

	for i, v := range ds.FeatureCollection.Features {
		points[i] = Record{v}
	}

	ds.Index = kdbush.NewBush(points, 10)
	
	return ds, nil
}


func (ds *DataSet) AddFeatures(features []*geojson.Feature) {
	ds.FeatureCollection.Features = append(ds.FeatureCollection.Features, features...)
}


func (ds *DataSet) GetFeature(i int) *geojson.Feature {
	return ds.FeatureCollection.Features[i];
}


func featureHandler(ds *DataSet, r *http.Request) ([]int, error) {
	bbox, err := UnmarshalBoundingBox(r.FormValue("bbox"))

	if err != nil {
		return nil, err
	} 

	return ds.Index.Range(bbox.MinX, bbox.MinY, bbox.MaxX, bbox.MaxY), nil
}


func nearestHandler(ds *DataSet, r *http.Request) ([]int, error) {
	point, err := UnmarshalPoint(r.FormValue("point"))
	if err != nil {
		return nil, err
	}

	radius, err := strconv.ParseFloat(r.FormValue("radius"), 10)
	if err != nil {
		return nil, err
	}

	return ds.Index.Within(point, radius), nil
}


func main() {
	files := make([]string, len(os.Args) - 1)
	copy(files, os.Args[1:])

	/* Read each GeoJSON file passed as argument */
	ds, err := NewDataSet(files)

	if err != nil {
		panic(err)
	}

	data = ds

	http.HandleFunc("/features", makeFeatureCollectionHandler(featureHandler));

	http.HandleFunc("/nearest", makeFeatureCollectionHandler(nearestHandler));

	http.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		ds, err := NewDataSet(files)

		if err == nil {
			data = ds
			w.WriteHeader(200)
			fmt.Fprintf(w, "%d features", len(data.FeatureCollection.Features))
		} else {
			w.WriteHeader(500)
			fmt.Fprintf(w, err.Error())
		}
	})
	
	http.ListenAndServe(":8000", nil);
}
