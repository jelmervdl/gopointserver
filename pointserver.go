package main

import (
	"fmt"
	"strings"
	"errors"
	"strconv"
	"os"
	"io/ioutil"
	"net/http"
	"github.com/MadAppGang/kdbush"
	"github.com/paulmach/go.geojson"
)


type Record struct {
	Feature *geojson.Feature
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


func makeFeatureCollectionHandler(fc *geojson.FeatureCollection, fn func(*http.Request) ([]int, error)) http.HandlerFunc {
	return makeJSONHandler(func(r *http.Request) ([]byte, error) {
		results, err := fn(r)
		if err != nil {
			return nil, err
		}

		rfc := geojson.NewFeatureCollection()

		for _, i := range results {
			rfc.AddFeature(fc.Features[i])
		}

		bytes, err := rfc.MarshalJSON()
		if err != nil {
			return nil, err
		}

		return bytes, nil
	})
}

func main() {
	fc := geojson.NewFeatureCollection()

	/* Read each GeoJSON file passed as argument */
	for _, path := range os.Args[1:] {
		dat, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		ffc, err := geojson.UnmarshalFeatureCollection(dat)
		if err != nil {
			panic(err)
		}

		fc.Features = append(fc.Features, ffc.Features...)
	}

	/* Create the index */

	points := make([]kdbush.Point, len(fc.Features))

	for i, v := range fc.Features {
		points[i] = Record{v}
	}

	fmt.Printf("Building index for %d records...\n", len(points))

	bush := kdbush.NewBush(points, 10)
	
	featureHandler := func(r *http.Request) ([]int, error) {
		bbox, err := UnmarshalBoundingBox(r.FormValue("bbox"))

		if err != nil {
			return nil, err
		} 

		return bush.Range(bbox.MinX, bbox.MinY, bbox.MaxX, bbox.MaxY), nil
	}

	http.HandleFunc("/features", makeFeatureCollectionHandler(fc, featureHandler));

	nearestHandler := func(r *http.Request) ([]int, error) {
		point, err := UnmarshalPoint(r.FormValue("point"))
		if err != nil {
			return nil, err
		}

		radius, err := strconv.ParseFloat(r.FormValue("radius"), 10)
		if err != nil {
			return nil, err
		}

		return bush.Within(point, radius), nil
	}

	http.HandleFunc("/nearest", makeFeatureCollectionHandler(fc, nearestHandler));
	
	http.ListenAndServe(":8000", nil);
}
