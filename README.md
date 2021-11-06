# asciimap
asciimap builds an ASCII map from a geojson file. The ascii map can be used for [astar](https://github.com/shanghuiyang/astar) project.

## usage
```
Application Options:
  -f, --geojson-file=FILENAME    Input geojson file name
  -g, --grid-size=GRIDSIZE       Grid size (default: 0.000010)
  -m, --map-file=MAPFILE         Output map file (default: map.txt)

Help Options:
  -h, --help                     Show this help message
```

### example
```shell
$asciimap -f map.geojson -g 0.00001 -m map.txt
```

### spec of geojson

