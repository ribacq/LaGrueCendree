// See LICENSE file for legal info
// Quentin RIBAC, december 2019
package main

// make your choice here
const (
	GRID_WIDTH  int    = 200
	GRID_HEIGHT int    = 100
	CONNECT_Y   bool   = false
	CONNECT_X   bool   = false
	FILENAME    string = "-"
)

func main() {
	terrain := GenerateTerrain()
	grid := AddFeaturesToTerrain(terrain)
	grid.DecorateFeatures()
	PrintPNG(grid)
}
