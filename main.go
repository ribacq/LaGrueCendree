// See LICENSE file for legal info
// Quentin RIBAC, december 2019
package main

import (
	"math/rand"
	"time"
)

// make your choice here
const (
	GRID_WIDTH  int  = 300
	GRID_HEIGHT int  = 150
	CONNECT_Y   bool = false
	CONNECT_X   bool = false
)

func main() {
	rand.Seed(time.Now().UnixNano())
	terrain := GenerateTerrain()
	grid := AddFeaturesToTerrain(terrain)
	grid.DecorateFeatures()
	PrintPNG(grid)
}
