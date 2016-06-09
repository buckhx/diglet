package mbt

import (
	"encoding/json"
	"io/ioutil"

	"github.com/buckhx/diglet/geo"
	"github.com/buckhx/diglet/util"
	"github.com/deckarep/golang-set"
	"github.com/kpawlik/geojson"
	//"github.com/buckhx/diglet/mbt/mvt"
	//"github.com/buckhx/diglet/mbt/mvt/vector_tile"
)

type GeojsonSource struct {
	path   string
	filter mapset.Set
}

func NewGeojsonSource(path string, filter []string) *GeojsonSource {
	var set mapset.Set
	if filter == nil {
		set = nil
	} else {
		set = mapset.NewSet()
		for _, k := range filter {
			set.Add(k)
		}
	}
	return &GeojsonSource{path, set}
}

func (gj *GeojsonSource) Publish(workers int) (features chan *geo.Feature, err error) {
	collection := readGeoJson(gj.path)
	f := make(chan *geo.Feature, 10)
	wg := util.NWork(func() {
		for _, feature := range collection.Features {
			f <- geojsonFeatureAdapter(feature)
		}
	}, workers)
	go func() {
		wg.Wait()
		close(features)
	}()
	features = f
	/*
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				for _, feature := range collection.Features {
					features <- geojsonFeatureAdapter(feature)
				}
			}()
		}
		go func() {
			wg.Wait()
			close(features)
		}()
	*/
	return
}

// Flatten all the points of a feature into single list. This can hel in identifying which tiles are going to be
// created
func geojsonFeatureAdapter(gj *geojson.Feature) (feature *geo.Feature) {
	// TODO: This sucks... I just want to switch on Coordinates.(type)
	igeom, err := gj.GetGeometry()
	util.Check(err)
	feature = geo.NewFeature(igeom.GetType())
	//TODO filter properties
	feature.Properties = gj.Properties
	feature.ID = gj.Id
	//TODO if id == nil assign a fake one
	feature.Type = igeom.GetType()
	switch geom := igeom.(type) {
	case *geojson.Point:
		shape := coordinatesAdapter(geojson.Coordinates{geom.Coordinates})
		feature.AddShape(shape)
	case *geojson.LineString:
		shape := coordinatesAdapter(geom.Coordinates)
		feature.AddShape(shape)
	case *geojson.MultiPoint:
		shape := coordinatesAdapter(geom.Coordinates)
		feature.AddShape(shape)
	case *geojson.MultiLineString:
		for _, line := range geom.Coordinates {
			shape := coordinatesAdapter(line)
			feature.AddShape(shape)
		}
	case *geojson.Polygon:
		// mvt need exterior ring to be clockwise
		// and interior rings to counter-clockwise
		exterior := true
		for _, line := range geom.Coordinates {
			shape := coordinatesAdapter(line)
			if exterior {
				if !shape.IsClockwise() {
					shape.Reverse()
				}
				exterior = false
			} else {
				if shape.IsClockwise() {
					shape.Reverse()
				}
			}
			feature.AddShape(shape)
		}
	case *geojson.MultiPolygon:
		for _, multiline := range geom.Coordinates {
			exterior := true
			for _, line := range multiline {
				shape := coordinatesAdapter(line)
				if exterior {
					if !shape.IsClockwise() {
						shape.Reverse()
					}
					exterior = false
				} else {
					if shape.IsClockwise() {
						shape.Reverse()
					}
				}
				feature.AddShape(shape)
			}
		}
	default:
		panic("Invalid Coordinate Type in GeoJson Feature") // + feature.String())
	}
	return
}
func coordinatesAdapter(line geojson.Coordinates) (shape *geo.Shape) {
	shape = geo.MakeShape(len(line))
	for i, point := range line {
		lat := float64(point[1])
		lon := float64(point[0])
		coord := geo.Coordinate{Lat: lat, Lon: lon}
		shape.Coordinates[i] = coord
	}
	return
}

func readGeoJson(path string) (features *geojson.FeatureCollection) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		util.Check(err)
	}
	if err := json.Unmarshal(file, &features); err != nil {
		util.Check(err)
	}
	return features
}

func publishFeatureCollection(collection *geojson.FeatureCollection) (features chan *geo.Feature) {
	features = make(chan *geo.Feature, 10)
	go func() {
		defer close(features)
		for _, feature := range collection.Features {
			features <- geojsonFeatureAdapter(feature)
		}
	}()
	return
}
