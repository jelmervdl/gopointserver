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

func (p Record) String() string {
	x, y := p.Coordinates()
	return fmt.Sprintf("%f %f", x, y)
}

type BoundingBox struct {
	MinX, MinY, MaxX, MaxY float64
}

func UnmarshalBoundingBox(str string) (BoundingBox, error) {
	components := strings.Split(str, ",")

	if len(components) != 4 {
		return BoundingBox{}, errors.New("Bbox string is not 4 components long")
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
	
	handler := func(w http.ResponseWriter, r *http.Request) {
		bbox, err := UnmarshalBoundingBox(r.FormValue("bbox"))

		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Error: %s", err)
			return
		} 

		results := bush.Range(bbox.MinX, bbox.MinY, bbox.MaxX, bbox.MaxY)
		rfc := geojson.NewFeatureCollection()

		for _, i := range results {
			rfc.AddFeature(fc.Features[i])
		}

		bytes, err := rfc.MarshalJSON()

		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "Error: %s", err)
			return
		}

		w.Header().Add("Content-Type", "application/json; charset=utf8")
		w.Header().Add("Content-Length", strconv.Itoa(len(bytes)))
		w.WriteHeader(200)
		w.Write(bytes)
	}

	http.HandleFunc("/", handler);
	http.ListenAndServe(":8000", nil);
}
