package main

import (
	"embed"
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/simplify"
	"github.com/ringsaturn/cnmaps-compress/convert"
	tzfrevert "github.com/ringsaturn/tzf/convert"
	"github.com/ringsaturn/tzf/pb"
	"github.com/twpayne/go-polyline"
	"google.golang.org/protobuf/proto"
)

//go:embed cnmaps/cnmaps/data/geojson.min/administrative/amap/land
var FS embed.FS

func CompressedPointsToPolylineBytes(points []*pb.Point) []byte {
	expect := [][]float64{}
	for _, point := range points {
		expect = append(expect, []float64{float64(point.Lng), float64(point.Lat)})
	}
	return polyline.EncodeCoords(expect)
}

func Compress(timezone *pb.Timezone) *pb.CompressedTimezone {
	reducedTimezone := &pb.CompressedTimezone{
		Name: timezone.Name,
	}
	for _, polygon := range timezone.Polygons {
		newPoly := &pb.CompressedPolygon{
			Points: CompressedPointsToPolylineBytes(polygon.Points),
			Holes:  make([]*pb.CompressedPolygon, 0),
		}
		for _, hole := range polygon.Holes {
			newPoly.Holes = append(newPoly.Holes, &pb.CompressedPolygon{
				Points: CompressedPointsToPolylineBytes(hole.Points),
			})
		}
		reducedTimezone.Data = append(reducedTimezone.Data, newPoly)
	}
	return reducedTimezone
}

func DecompressedPolylineBytesToPoints(input []byte) []*pb.Point {
	expect := []*pb.Point{}
	coords, _, _ := polyline.DecodeCoords(input)
	for _, coord := range coords {
		expect = append(expect, &pb.Point{
			Lng: float32(coord[0]), Lat: float32(coord[1]),
		})
	}
	return expect
}

func Decompress(timezone *pb.CompressedTimezone) *pb.Timezone {
	reducedTimezone := &pb.Timezone{
		Name: timezone.Name,
	}
	for _, polygon := range timezone.Data {
		newPoly := &pb.Polygon{
			Points: DecompressedPolylineBytesToPoints(polygon.Points),
			Holes:  make([]*pb.Polygon, 0),
		}
		for _, hole := range polygon.Holes {
			newPoly.Holes = append(newPoly.Holes, &pb.Polygon{
				Points: DecompressedPolylineBytesToPoints(hole.Points),
			})
		}
		reducedTimezone.Polygons = append(reducedTimezone.Polygons, newPoly)
	}
	return reducedTimezone
}

func ReducePoints(points []*pb.Point) []*pb.Point {
	if len(points) == 0 {
		return points
	}
	original := orb.LineString{}
	for _, point := range points {
		original = append(original, orb.Point{float64(point.Lng), float64(point.Lat)})
	}
	reduced := simplify.DouglasPeucker(0.001).Simplify(original.Clone()).(orb.LineString)
	res := make([]*pb.Point, 0)
	for _, orbPoint := range reduced {
		res = append(res, &pb.Point{
			Lng: float32(orbPoint.Lon()),
			Lat: float32(orbPoint.Lat()),
		})
	}
	return res
}

func Redcude(timezone *pb.Timezone) *pb.Timezone {
	reducedTimezone := &pb.Timezone{
		Name: timezone.Name,
	}
	for _, polygon := range timezone.Polygons {
		newPoly := &pb.Polygon{
			Points: ReducePoints(polygon.Points),
			Holes:  make([]*pb.Polygon, 0),
		}
		for _, hole := range polygon.Holes {
			newPoly.Holes = append(newPoly.Holes, &pb.Polygon{
				Points: ReducePoints(hole.Points),
			})
		}
		reducedTimezone.Polygons = append(reducedTimezone.Polygons, newPoly)
	}
	return reducedTimezone
}

func main() {
	entries, err := FS.ReadDir("cnmaps/cnmaps/data/geojson.min/administrative/amap/land")
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".geojson") {
			continue
		}

		originalPath := "cnmaps/cnmaps/data/geojson.min/administrative/amap/land/" + entry.Name()
		content, err := FS.ReadFile(originalPath)
		if err != nil {
			panic(err)
		}

		d := &convert.GeometryDefine{}
		if err := json.Unmarshal(content, d); err != nil {
			panic(err)
		}

		outputpb := Decompress(Compress(Redcude(convert.GeometryDefineToTZPB(d))))
		output := tzfrevert.RevertItem(outputpb)

		func() {
			outputPath := strings.Replace(originalPath, ".geojson", ".json", 1)
			outputBin, err := json.Marshal(output)
			if err != nil {
				panic(err)
			}
			_ = ioutil.WriteFile(outputPath, outputBin, 0644)
		}()

		func() {
			outputPath := strings.Replace(originalPath, ".geojson", ".pb", 1)
			outputBin, err := proto.Marshal(outputpb)
			if err != nil {
				panic(err)
			}
			_ = ioutil.WriteFile(outputPath, outputBin, 0644)
		}()

	}
}
