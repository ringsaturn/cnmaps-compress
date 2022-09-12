package convert

import "github.com/ringsaturn/tzf/pb"

const (
	MultiPolygonType = "MultiPolygon"
	PolygonType      = "Polygon"
	FeatureType      = "Feature"
)

type PolygonCoordinates [][][2]float64
type MultiPolygonCoordinates []PolygonCoordinates

type GeometryDefine struct {
	Coordinates MultiPolygonCoordinates `json:"coordinates"`
	Type        string                  `json:"type"`
	Properties  PropertiesDefine        `json:"properties"`
}

type PropertiesDefine struct {
	Tzid string `json:"adcode"`
}

func GeometryDefineToTZPB(d *GeometryDefine) *pb.Timezone {
	pbtz := &pb.Timezone{}
	coordinates := d.Coordinates

	polygons := make([]*pb.Polygon, 0)
	for _, subcoordinates := range coordinates {
		newpbPoly := &pb.Polygon{
			Points: make([]*pb.Point, 0),
			Holes:  make([]*pb.Polygon, 0),
		}
		for index, geoPoly := range subcoordinates {
			if index == 0 {
				for _, rawCoords := range geoPoly {
					newpbPoly.Points = append(newpbPoly.Points, &pb.Point{
						Lng: float32(rawCoords[0]),
						Lat: float32(rawCoords[1]),
					})
				}
				continue
			}

			holePoly := &pb.Polygon{
				Points: make([]*pb.Point, 0),
			}
			for _, rawCoords := range geoPoly {
				holePoly.Points = append(holePoly.Points, &pb.Point{
					Lng: float32(rawCoords[0]),
					Lat: float32(rawCoords[1]),
				})
			}
			newpbPoly.Holes = append(newpbPoly.Holes, holePoly)

		}
		polygons = append(polygons, newpbPoly)
	}
	pbtz.Polygons = polygons
	return pbtz
}
