package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/geo/s2"
	geojson "github.com/paulmach/go.geojson"
	"github.com/shanghuiyang/astar/tilemap"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	defaultGridSize = "0.00001"
	defaultMapFile  = "map.txt"
)

var (
	mapbbox  *bbox
	gridSize = 0.00001
)

type bbox struct {
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
	Top    float64 `json:"top"`
	Bottom float64 `json:"buttom"`
}

func main() {

	geojsonFile := kingpin.Arg("geojson-file", "Input geojson file name(required)").Required().String()
	gz := kingpin.Flag("grid-size", "Grid size").Short('g').Default(defaultGridSize).Float64()
	mapFile := kingpin.Flag("map-file", "Output map file").Short('m').Default(defaultMapFile).String()
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	gridSize = *gz
	rawJSON, err := loadMap(*geojsonFile)
	if err != nil {
		fmt.Printf("ERROR! %v\n", err)
		os.Exit(1)
	}

	fc, err := geojson.UnmarshalFeatureCollection([]byte(rawJSON))
	if err != nil {
		fmt.Printf("ERROR! %v\n", err)
		os.Exit(1)
	}

	var bb *geojson.Geometry
	var walls []*s2.Loop
	for _, f := range fc.Features {
		if !f.Geometry.IsPolygon() {
			fmt.Printf("WARNING! skip a geometry who isn't polygon\n")
			continue
		}
		if isbbox, _ := f.PropertyBool("isbbox"); isbbox {
			bb = f.Geometry
			continue
		}
		w := toLoop(f.Geometry)
		walls = append(walls, w)
	}

	mapbbox = bbound(bb)
	row := int((mapbbox.Top-mapbbox.Bottom)/gridSize + 0.5)
	col := int((mapbbox.Right-mapbbox.Left)/gridSize + 0.5)
	fmt.Println(row, col)
	m := tilemap.New(row, col)
	for r := 0; r < row; r++ {
		for c := 0; c < col; c++ {
			lat, lon := xy2latlon(r, c)
			pt := s2.PointFromLatLng(s2.LatLngFromDegrees(lat, lon))
			for _, w := range walls {
				if w.ContainsPoint(pt) {
					m.SetWall(r, c)
				}
			}
		}
	}

	mapstr := m.String()
	ioutil.WriteFile(*mapFile, []byte(mapstr), os.ModePerm)
	fmt.Printf("map bbox:\n")
	fmt.Printf(" left  : %11.6f\n", mapbbox.Left)
	fmt.Printf(" right : %11.6f\n", mapbbox.Right)
	fmt.Printf(" top   : %11.6f\n", mapbbox.Top)
	fmt.Printf(" bottom: %11.6f\n", mapbbox.Bottom)
	fmt.Printf("grid size: %.6f\n", gridSize)
	fmt.Printf("%v", mapstr)
	os.Exit(0)
}

func xy2latlon(x, y int) (lat, lon float64) {
	return mapbbox.Top - float64(x)*gridSize, mapbbox.Left + float64(y)*gridSize
}

func loadMap(file string) ([]byte, error) {
	rawJSON, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err

	}
	return rawJSON, nil
}

func bbound(g *geojson.Geometry) *bbox {
	b := &bbox{
		Left:   181,
		Right:  -181,
		Top:    -91,
		Bottom: 91,
	}
	for _, poly := range g.Polygon {
		for _, pt := range poly {
			if pt[1] < b.Bottom {
				b.Bottom = pt[1]
			}
			if pt[1] > b.Top {
				b.Top = pt[1]
			}
			if pt[0] < b.Left {
				b.Left = pt[0]
			}
			if pt[0] > b.Right {
				b.Right = pt[0]
			}
		}
	}
	return b
}

func toLoop(g *geojson.Geometry) *s2.Loop {
	if !g.IsPolygon() {
		return nil
	}
	if len(g.Polygon) == 0 {
		return nil
	}
	var pts []s2.Point
	for i := 0; i < len(g.Polygon[0])-1; i++ {
		pt := g.Polygon[0][i]
		pts = append(pts, s2.PointFromLatLng(s2.LatLngFromDegrees(pt[1], pt[0])))
	}
	return s2.LoopFromPoints(pts)
}
