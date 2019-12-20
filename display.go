package main

import (
	"fmt"
	"math/rand"
)

const (
	TERRAIN_SEA = iota
	TERRAIN_LAND
	TERRAIN_MOUNTAIN
	TERRAIN_RIVER
	TERRAIN_RIVER_HEAD
	TERRAIN_CITY
	TERRAIN_MAP_BORDER
)

var (
	DIR_NEXT [4][2]int = [4][2]int{
		{0, 1}, {1, 0}, {0, -1}, {-1, 0},
	}
	DIR_SQUARE [8][2]int = [8][2]int{
		{0, 1}, {1, 0}, {0, -1}, {-1, 0},
		{-1, -1}, {-1, 1}, {1, -1}, {1, 1},
	}
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
	Y, X int
	Dir  int
}

type City struct {
	Y, X int
	Size int
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
		if se >= nLand*5/100 {
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
	for len(rivers) < MAGIC/3 {
		y, x := rand.Intn(GRID_HEIGHT), rand.Intn(GRID_WIDTH)
		if grid[y][x].Terrain != TERRAIN_SEA {
			grid[y][x].Terrain = TERRAIN_RIVER_HEAD
			rivers = append(rivers, &River{
				Y:   y,
				X:   x,
				Dir: rand.Intn(len(DIR_NEXT)),
			})
		}
	}
	for len(rivers) > 0 {
		river := rivers[0]
		// for each river still alive, decide where to go
		highDir, highLevel := -1, -1
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
			continue
		} else if len(rivers) > 1 {
			rivers = rivers[1:]
		} else {
			rivers = []*River{}
		}
	}

	// cities
	var cities []*City
	for len(cities) < MAGIC {
		y, x := rand.Intn(GRID_HEIGHT), rand.Intn(GRID_WIDTH)
		if grid[y][x].Terrain != TERRAIN_LAND {
			continue
		}

		city := &City{
			Y: y,
			X: x,
		}

		cityUpsizers := map[int]bool{
			TERRAIN_SEA:        false,
			TERRAIN_RIVER:      false,
			TERRAIN_RIVER_HEAD: false,
		}
		for _, dir := range DIRECTIONS {
			for terrain := range cityUpsizers {
				nhbY, nhbX := Inside(y+dir[0], x+dir[1])
				if grid[nhbY][nhbX].Terrain == terrain {
					cityUpsizers[terrain] = true
				}
			}
		}
		for _, ok := range cityUpsizers {
			if ok {
				city.Size++
			}
		}

		if city.Size == 0 && rand.Intn(100) < 90 {
			continue
		}

		grid[y][x].Terrain = TERRAIN_CITY
		switch city.Size {
		case 0:
			grid[y][x].Color = Color{0, 0, 0}
		case 1:
			grid[y][x].Color = Color{255, 255, 255}
		case 2:
			grid[y][x].Color = Color{255, 255, 255}
			for _, dir := range DIR_SQUARE {
				nhbY, nhbX := Inside(y+dir[0], x+dir[1])
				if grid[nhbY][nhbX].Terrain == TERRAIN_LAND {
					grid[nhbY][nhbX].Terrain = TERRAIN_CITY
					grid[nhbY][nhbX].Color = Color{232, 119, 13}
				}
			}
		case 3:
			grid[y][x].Color = Color{255, 255, 255}
			for _, dir := range DIRECTIONS {
				nhbY, nhbX := Inside(y+dir[0], x+dir[1])
				if grid[nhbY][nhbX].Terrain == TERRAIN_LAND {
					grid[nhbY][nhbX].Terrain = TERRAIN_CITY
					grid[nhbY][nhbX].Color = Color{232, 13, 17}
				}
			}
		}
		cities = append(cities, city)
	}

	// map borders
	if !CONNECT_Y {
		for x := 0; x < GRID_WIDTH; x++ {
			grid[0][x].Terrain = TERRAIN_MAP_BORDER
			grid[GRID_HEIGHT-1][x].Terrain = TERRAIN_MAP_BORDER
			if x%MAGIC < MAGIC/2 {
				grid[0][x].Color = Color{45, 45, 45}
				grid[GRID_HEIGHT-1][x].Color = Color{150, 150, 150}
			} else {
				grid[0][x].Color = Color{150, 150, 150}
				grid[GRID_HEIGHT-1][x].Color = Color{45, 45, 45}
			}
		}
	}
	if !CONNECT_X {
		for y := 0; y < GRID_HEIGHT; y++ {
			grid[y][0].Terrain = TERRAIN_MAP_BORDER
			grid[y][GRID_WIDTH-1].Terrain = TERRAIN_MAP_BORDER
			if y%MAGIC < MAGIC/2 {
				grid[y][0].Color = Color{150, 150, 150}
				grid[y][GRID_WIDTH-1].Color = Color{45, 45, 45}
			} else {
				grid[y][0].Color = Color{45, 45, 45}
				grid[y][GRID_WIDTH-1].Color = Color{150, 150, 150}
			}
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
					R: v * 224 / 255,
					G: v * 228 / 255,
					B: v * 170 / 255,
				}
			case TERRAIN_RIVER, TERRAIN_RIVER_HEAD:
				grid[y][x].Color = Color{
					R: v / 4,
					G: v / 2,
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
