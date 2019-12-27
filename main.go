// See LICENSE file for legal info
// Quentin RIBAC, december 2019
package main

import (
	"math/rand"
	"time"
)

// make your choice here
const (
	GRID_WIDTH  int  = 200
	GRID_HEIGHT int  = 100
	CONNECT_Y   bool = true
	CONNECT_X   bool = true
)

func main() {
	rand.Seed(time.Now().UnixNano())
	terrain := GenerateTerrain()
	grid := AddFeaturesToTerrain(terrain)
	grid.DecorateFeatures()
	PrintPNG(grid)
}
