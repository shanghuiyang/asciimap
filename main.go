package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/shanghuiyang/astar"
	"github.com/shanghuiyang/astar/tilemap"
	"github.com/spatial-go/geoos/geojson"
	"github.com/spatial-go/geoos/planar"
	"github.com/spatial-go/geoos/space"
)

var (
	mapbbox  *bbox
	gridSize = 0.00001
)

type point struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type bbox struct {
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
	Top    float64 `json:"top"`
	Bottom float64 `json:"buttom"`
}

func main() {
	var opts struct {
		GeojsonFile string  `short:"f" long:"geojson-file" required:"true" description:"Input geojson file nam(required)" value-name:"FILENAME"`
		GridSize    float64 `short:"g" long:"grid-size" description:"Grid size" value-name:"GRIDSIZE" default:"0.00001"`
		MapFile     string  `short:"m" long:"map-file" description:"Output map file" value-name:"MAPFILE" default:"map.txt"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	gridSize = opts.GridSize
	rawJSON, err := loadMap(opts.GeojsonFile)
	if err != nil {
		fmt.Printf("ERROR! %v\n", err)
		os.Exit(1)
	}

	strategy := planar.NormalStrategy()
	fc, err := geojson.UnmarshalFeatureCollection([]byte(rawJSON))
	if err != nil {
		fmt.Printf("ERROR! %v\n", err)
		os.Exit(1)
	}
	var bb space.Geometry
	var walls []space.Geometry
	for _, f := range fc.Features {
		if v, ok := f.Properties["isbbox"]; ok && v.(bool) {
			bb = f.Geometry.Coordinates
			continue
		}
		walls = append(walls, f.Geometry.Coordinates)
	}

	b := bb.Bound()
	mapbbox = &bbox{
		Left:   b.Min.Lon(),
		Right:  b.Max.Lon(),
		Top:    b.Max.Lat(),
		Bottom: b.Min.Lat(),
	}

	row := int((mapbbox.Top-mapbbox.Bottom)/gridSize + 0.5)
	col := int((mapbbox.Right-mapbbox.Left)/gridSize + 0.5)
	m := tilemap.New(row, col)
	for r := 0; r < row; r++ {
		for c := 0; c < col; c++ {
			pt := xy2geo(&astar.Point{X: r, Y: c})
			p := &space.Point{pt.Lon, pt.Lat}
			for _, w := range walls {
				yes, err := strategy.Intersects(p, w)
				if err != nil {
					fmt.Printf("ERROR! %v\n", err)
					os.Exit(1)
				}
				if yes {
					m.SetWall(r, c)
				}
			}
		}
	}

	mapstr := m.String()
	ioutil.WriteFile(opts.MapFile, []byte(mapstr), os.ModePerm)
	fmt.Printf("map bbox:\n")
	fmt.Printf("-left:\t\t%11.6f\n", mapbbox.Left)
	fmt.Printf("-right:\t\t%11.6f\n", mapbbox.Right)
	fmt.Printf("-top:\t\t%11.6f\n", mapbbox.Top)
	fmt.Printf("-bottom:\t%11.6f\n", mapbbox.Bottom)
	fmt.Printf("grid size: \t%11.6f\n", gridSize)
	fmt.Printf("%v", mapstr)
	os.Exit(0)
}

func xy2geo(p *astar.Point) *point {
	return &point{
		Lat: mapbbox.Top - float64(p.X)*gridSize,
		Lon: mapbbox.Left + float64(p.Y)*gridSize,
	}
}

func loadMap(file string) ([]byte, error) {
	rawJSON, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err

	}
	return rawJSON, nil
}
