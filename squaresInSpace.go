// See LICENSE file for legal info
// Quentin RIBAC, december 2019
package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
)

// make your choice here
const (
	GRID_WIDTH  int  = 256
	GRID_HEIGHT int  = 128
	CONNECT_Y   bool = false
	CONNECT_X   bool = true
	GIF         bool = false
)

// handle with care
var (
	SURFACE     int = GRID_WIDTH * GRID_HEIGHT
	DIAGONAL    int = GRID_WIDTH + GRID_HEIGHT
	MAGIC       int = int(2.0 * math.Sqrt(float64(DIAGONAL)))
	MAX_VAL     int = MAGIC
	FRAMES      int = MAGIC
	SPAWN_POWER int = MAGIC
)

var (
	DIRECTIONS [20][2]int = [20][2]int{
		{0, 1}, {0, -1}, {1, 0}, {-1, 0},
		{-1, -1}, {-1, 1}, {1, -1}, {1, 1},
		{0, 2}, {0, -2}, {2, 0}, {-2, 0},
		{-2, -1}, {-2, 1}, {-1, 2}, {1, 2},
		{2, 1}, {2, -1}, {1, -2}, {-1, -2},
	}
	DIR_NEXT [4][2]int = [4][2]int{
		{0, 1}, {1, 0}, {0, -1}, {-1, 0},
	}
)

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Sign(x int) int {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	}
	return 0
}

func Inside(y, x int) (int, int) {
	for y < 0 {
		y += GRID_HEIGHT
	}
	for y >= GRID_HEIGHT {
		y -= GRID_HEIGHT
	}
	for x < 0 {
		x += GRID_WIDTH
	}
	for x >= GRID_WIDTH {
		x -= GRID_WIDTH
	}
	return y, x
}

const (
	TERRAIN_SEA = iota
	TERRAIN_LAND
	TERRAIN_MOUNTAIN
	TERRAIN_RIVER
	TERRAIN_RIVER_HEAD
)

type Color struct {
	R, G, B int
}

type SquareTerrain struct {
	Val          int
	Terrain      int
	Surroundings int
	Color        Color
}

type River struct {
	Y, X  int
	Dir   int
	Level int
}

func display(squares [GRID_HEIGHT][GRID_WIDTH]int) (nLand, nSea int) {
	// inner model
	var grid [GRID_HEIGHT][GRID_WIDTH]SquareTerrain
	for y := range grid {
		for x := range grid[y] {
			grid[y][x].Val = squares[y][x]
		}
	}

	// delete isolated
	for i := 0; i < 2; i++ {
		for y := range grid {
			for x := range grid[y] {
				grid[y][x].Surroundings = 0
				for dir := range DIRECTIONS {
					nhbY, nhbX := Inside(y+DIRECTIONS[dir][0], x+DIRECTIONS[dir][1])
					grid[y][x].Surroundings += grid[nhbY][nhbX].Val
				}
				if grid[y][x].Surroundings*grid[y][x].Val < 0 {
					grid[y][x].Val *= -1
				}
			}
		}
	}

	// terrain: land and sea
	for y := range grid {
		for x := range grid[y] {
			if grid[y][x].Val <= 0 {
				grid[y][x].Terrain = TERRAIN_SEA
				grid[y][x].Val *= -1
				nSea++
			} else {
				grid[y][x].Terrain = TERRAIN_LAND
				nLand++
			}
		}
	}

	// smooth and detect min and max land and sea
	minL, maxL, minS, maxS := -1, -1, -1, -1
	for _, y := range rand.Perm(len(grid)) {
		for _, x := range rand.Perm(len(grid[y])) {
			for dir := range DIRECTIONS {
				nhbY, nhbX := Inside(y+DIRECTIONS[dir][0], x+DIRECTIONS[dir][1])
				grid[y][x].Val += grid[nhbY][nhbX].Val
			}
			grid[y][x].Val /= 1 + len(DIRECTIONS)

			// normalize
			if grid[y][x].Val >= MAX_VAL {
				grid[y][x].Val = MAX_VAL - 1
			}

			// min and max land
			if grid[y][x].Terrain == TERRAIN_LAND {
				if minL == -1 || minL > grid[y][x].Val {
					minL = grid[y][x].Val
				}
				if maxL == -1 || maxL < grid[y][x].Val {
					maxL = grid[y][x].Val
				}
			}

			// min and max sea
			if grid[y][x].Terrain == TERRAIN_SEA {
				if minS == -1 || minS > grid[y][x].Val {
					minS = grid[y][x].Val
				}
				if maxS == -1 || maxS < grid[y][x].Val {
					maxS = grid[y][x].Val
				}
			}
		}
	}

	// normalize values to 255
	for y := range grid {
		for x := range grid[y] {
			if grid[y][x].Terrain == TERRAIN_LAND {
				grid[y][x].Val = grid[y][x].Val * 255 / maxL
			} else if grid[y][x].Terrain == TERRAIN_SEA {
				grid[y][x].Val = grid[y][x].Val * 255 / maxS
			}
		}
	}

	// evelation map
	var elevation [256]int
	for y := range grid {
		for x := range grid[y] {
			if grid[y][x].Terrain == TERRAIN_LAND {
				elevation[grid[y][x].Val]++
			}
		}
	}

	// mountains
	var se, maxEl int
	for el, count := range elevation {
		se += count
		if se >= nLand*3/100 {
			maxEl = el
			break
		}
	}
	for y := range grid {
		for x := range grid[y] {
			if grid[y][x].Terrain == TERRAIN_LAND && grid[y][x].Val <= maxEl {
				grid[y][x].Terrain = TERRAIN_MOUNTAIN
			}
		}
	}

	// rivers
	var rivers []*River
	for y := range grid {
		for x := range grid[y] {
			if grid[y][x].Terrain != TERRAIN_SEA && rand.Intn(DIAGONAL) < 1 && len(rivers) < MAGIC/5 {
				grid[y][x].Terrain = TERRAIN_RIVER_HEAD
				rivers = append(rivers, &River{
					Y:     y,
					X:     x,
					Dir:   rand.Intn(len(DIR_NEXT)),
					Level: grid[y][x].Val,
				})
			}
		}
	}
	for len(rivers) > 0 {
		river := rivers[0]
		// for each river still alive, decide where to go
		highDir, highLevel := -1, river.Level
		highLDir := -1
		var nhbY, nhbX int
		lake := true
		atSea := false
		dir := river.Dir
		for _, offset := range rand.Perm(len(DIR_NEXT)) {
			dir = (dir + offset) % len(DIR_NEXT)
			if dir == (river.Dir+2)%len(DIR_NEXT) {
				continue
			}
			nhbY, nhbX = Inside(river.Y+DIR_NEXT[dir][0], river.X+DIR_NEXT[dir][1])
			if grid[nhbY][nhbX].Terrain != TERRAIN_RIVER {
				lake = false
			}
			if grid[nhbY][nhbX].Terrain == TERRAIN_SEA {
				highDir = dir
				atSea = true
				break
			}
			if grid[nhbY][nhbX].Val > highLevel {
				highDir = dir
				highLevel = grid[nhbY][nhbX].Val
				if grid[nhbY][nhbX].Terrain != TERRAIN_RIVER {
					highLDir = highDir
				}
			}
		}

		// strengthen
		river.Level--

		// fill up and go somewhere is possible
		if highLDir != -1 && !atSea {
			highDir = highLDir
		}
		if highDir == -1 {
			continue
		}
		nhbY, nhbX = Inside(river.Y+DIR_NEXT[highDir][0], river.X+DIR_NEXT[highDir][1])
		if grid[nhbY][nhbX].Terrain != TERRAIN_SEA || lake {
			if grid[nhbY][nhbX].Terrain == TERRAIN_RIVER {
				grid[river.Y][river.X].Terrain = TERRAIN_LAND
			} else {
				grid[river.Y][river.X].Terrain = TERRAIN_RIVER
			}
			grid[nhbY][nhbX].Terrain = TERRAIN_RIVER_HEAD
			river.Y = nhbY
			river.X = nhbX
		} else if len(rivers) > 1 {
			rivers = rivers[1:]
		} else {
			rivers = []*River{}
		}
	}

	// colors
	for y := range grid {
		for x := range grid[y] {
			switch v := grid[y][x].Val; grid[y][x].Terrain {
			case TERRAIN_LAND:
				grid[y][x].Color = Color{
					R: v / 2,
					G: v,
					B: 0,
				}
			case TERRAIN_MOUNTAIN:
				grid[y][x].Color = Color{
					R: v * 248 / 255,
					G: v * 200 / 255,
					B: v * 120 / 255,
				}
			case TERRAIN_RIVER_HEAD:
				grid[y][x].Color = Color{
					R: 255,
					G: 255,
					B: 255,
				}
				fallthrough
			case TERRAIN_RIVER:
				grid[y][x].Color = Color{
					R: 0,
					G: 0,
					B: v,
				}
			case TERRAIN_SEA:
				grid[y][x].Color = Color{
					R: 0,
					G: 0,
					B: 255 - v*3/4,
				}
			}
		}
	}

	// display
	fmt.Println("P3")
	fmt.Println(GRID_WIDTH, GRID_HEIGHT, 255)
	for y := range grid {
		for x := range grid[y] {
			fmt.Print(grid[y][x].Color.R, grid[y][x].Color.G, grid[y][x].Color.B, " ")
		}
		fmt.Println()
	}

	return
}

func main() {
	var squares [GRID_HEIGHT][GRID_WIDTH]int
	println(GRID_WIDTH, "x", GRID_HEIGHT, "f", FRAMES)

	// spawn
	newY, newX := rand.Intn(GRID_HEIGHT), rand.Intn(GRID_WIDTH)
	for i := 0; i < SPAWN_POWER; i++ {
		if squares[newY][newX] == 0 {
			squares[newY][newX] = MAX_VAL
		}
		newY, newX = Inside(newY+rand.Intn(MAX_VAL), newX+rand.Intn(MAX_VAL))
	}

	// main loop on frames
	var wg sync.WaitGroup
	for frame := 1; frame <= FRAMES; frame++ {
		// loop on squares
		sy := 0
		for _, y := range rand.Perm(GRID_HEIGHT) {
			sy++
			print("\rframe ", frame, " (", sy, ")")
			for _, x := range rand.Perm(GRID_WIDTH) {
				wg.Add(1)
				go func(y, x int) {
					defer wg.Done()

					// skip empty
					if squares[y][x] == 0 {
						return
					}

					// correct excess negatives
					if squares[y][x] <= -MAX_VAL {
						squares[y][x] = 1 - MAX_VAL
					}

					// move to other
					dy, dx := 0, 0
					dist := ((squares[y][x]*2 + MAX_VAL) % MAX_VAL)
					for yo := y - dist; yo <= y+dist; yo++ {
						for xo := x - dist; xo <= x+dist; xo++ {
							// yo, xo inside
							inyo, inxo := Inside(yo, xo)

							// skip empty and far away
							if squares[inyo][inxo]*squares[y][x] <= 0 || (yo-y)*(yo-y)+(xo-x)*(xo-x) >= dist*dist {
								continue
							}

							// count force
							if CONNECT_Y || inyo == yo {
								dy += (yo - y)
							}
							if CONNECT_X || inxo == xo {
								dx += (xo - x)
							}
						}
					}
					ry, rx := Abs(dy), Abs(dx)
					if squares[y][x] < 0 {
						dy /= -squares[y][x]
						dx /= -squares[y][x]
					} else if dist > 0 {
						dy /= dist
						dx /= dist
					}
					nextY, nextX := Inside(y+dy, x+dx)
					for (dy != 0 || dx != 0) && squares[nextY][nextX]*squares[y][x] > 0 {
						if rand.Intn(ry+rx) < ry {
							dy -= Sign(dy)
						} else {
							dx -= Sign(dx)
						}
						nextY, nextX = Inside(y+dy, x+dx)
					}
					if squares[nextY][nextX]*squares[y][x] <= 0 {
						squares[nextY][nextX], squares[y][x] = squares[y][x]-Sign(squares[y][x]), squares[nextY][nextX]-Sign(squares[y][x])
						return
					}

					// move randomly
					for dir := range rand.Perm(len(DIRECTIONS)) {
						nextY, nextX = Inside(y+DIRECTIONS[dir][0], x+DIRECTIONS[dir][1])
						if squares[nextY][nextX]*squares[y][x] <= 0 {
							squares[nextY][nextX], squares[y][x] = squares[y][x]+Sign(squares[y][x]), squares[nextY][nextX]-Sign(squares[y][x])
							return
						}
					}
				}(y, x)
			}
		}
		wg.Wait()

		// display
		if GIF {
			print("\tgenerating image... ")
			display(squares)
			print("done")
		}
	}

	// end
	print("\ngenerating image...")
	nLand, nSea := display(squares)
	println("done\nLand:", nLand, "Sea:", nSea, "Land%:", 100*nLand/SURFACE)
}
