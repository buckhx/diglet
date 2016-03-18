package geo

import (
	"encoding/json"
	"github.com/buckhx/diglet/util"
	"github.com/deckarep/golang-set"
	"github.com/kpawlik/geojson"
	"io/ioutil"
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

func (gj *GeojsonSource) Publish() (features chan *Feature, err error) {
	collection := readGeoJson(gj.path)
	return publishFeatureCollection(collection), nil
}

// Flatten all the points of a feature into single list. This can hel in identifying which tiles are going to be
// created
func GeojsonFeatureAdapter(gj *geojson.Feature) (feature *Feature, err error) {
	// TODO: This sucks... I just want to switch on Coordinates.(type)
	igeom, err := gj.GetGeometry()
	if igeom == nil || err != nil {
		err = util.Errorf("Invalid geojson feature %q", gj)
		return
	}
	feature = NewFeature(igeom.GetType())
	//TODO filter properties
	feature.Properties = gj.Properties
	if feature.Properties != nil {
		feature.Properties["id"] = gj.Id
	}
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
		feature = nil

		err = util.Errorf("Invalid Coordinate Type in GeoJson %q", geom)
	}
	return
}

func UnmarshalGeojsonFeature(raw string) (feature *geojson.Feature, err error) {
	err = json.Unmarshal([]byte(raw), &feature)
	return
}

func coordinatesAdapter(line geojson.Coordinates) (shape *Shape) {
	shape = MakeShape(len(line))
	for i, point := range line {
		lat := float64(point[1])
		lon := float64(point[0])
		coord := Coordinate{Lat: lat, Lon: lon}
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

func publishFeatureCollection(collection *geojson.FeatureCollection) (features chan *Feature) {
	features = make(chan *Feature, 10)
	go func() {
		defer close(features)
		for _, feature := range collection.Features {
			f, err := GeojsonFeatureAdapter(feature)
			util.Warn(err, "feature publishing")
			features <- f
		}
	}()
	return
}
