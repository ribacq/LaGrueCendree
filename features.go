package main

import (
	"image/color"
	"math/rand"
)

const (
	TERRAIN_SEA = iota
	TERRAIN_LAND
	TERRAIN_MOUNTAIN
	TERRAIN_MAP_BORDER
)

const (
	FEATURE_NONE = iota
	FEATURE_RIVER
	FEATURE_CITY
	FEATURE_COUNTRY_BORDER
)

var (
	RIVER_PCT    int = 8
	NB_CITIES    int = MAGIC
	NB_COUNTRIES int = MAGIC / 5
)

const (
	SQUARE_WIDTH  int = 8
	SQUARE_HEIGHT int = 8
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

func HSVtoRGBA(H int, S, V float64) (out color.RGBA) {
	for H < 0 {
		H += 360
	}
	for H > 360 {
		H -= 360
	}
	C := V * S
	X := C * float64(1-Abs(H/60%2-1))
	m := V - C
	var r, g, b float64
	if H < 60 {
		r, g, b = C, X, 0.0
	} else if H < 120 {
		r, g, b = X, C, 0.0
	} else if H < 180 {
		r, g, b = 0.0, C, X
	} else if H < 240 {
		r, g, b = 0.0, X, C
	} else if H < 300 {
		r, g, b = X, 0.0, C
	} else {
		r, g, b = C, 0.0, X
	}
	out = color.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
	}
	if out.R > 255 {
		out.R = 255
	} else if out.R < 0 {
		out.R = 0
	}
	if out.G > 255 {
		out.G = 255
	} else if out.G < 0 {
		out.G = 0
	}
	if out.B > 255 {
		out.B = 255
	} else if out.B < 0 {
		out.B = 0
	}
	return
}

type Grid [GRID_HEIGHT][GRID_WIDTH]*SquareTerrain

type SquareTerrain struct {
	Val          int
	Terrain      int
	Feature      int
	Colors       [SQUARE_HEIGHT][SQUARE_WIDTH]color.Color
	CountryIndex int
}

func (st *SquareTerrain) SetRGBA(r, g, b, a uint8) {
	for y := range st.Colors {
		for x := range st.Colors[y] {
			st.Colors[y][x] = color.RGBA{r, g, b, a}
		}
	}
}

func (st *SquareTerrain) SetColor(c color.Color) {
	for y := range st.Colors {
		for x := range st.Colors[y] {
			st.Colors[y][x] = c
		}
	}
}

type River struct {
	y, x      []int
	pathStack []int
	Level     int
}

func NewRiver(grid *Grid) *River {
	y, x := rand.Intn(GRID_HEIGHT), rand.Intn(GRID_WIDTH)
	for grid[y][x].Terrain != TERRAIN_MOUNTAIN {
		y, x = rand.Intn(GRID_HEIGHT), rand.Intn(GRID_WIDTH)
	}
	grid[y][x].Feature = FEATURE_RIVER
	return &River{
		y:         []int{y},
		x:         []int{x},
		pathStack: []int{0},
		Level:     int(grid[y][x].Val),
	}
}

func (r *River) Y() int {
	return r.y[r.i()]
}

func (r *River) X() int {
	return r.x[r.i()]
}

func (r *River) i() int {
	return r.pathStack[len(r.pathStack)-1]
}

func (r *River) Len() int {
	return len(r.pathStack)
}

func (r *River) IsAt(y, x int) bool {
	for _, i := range r.pathStack {
		if r.y[i] == y && r.x[i] == x {
			return true
		}
	}
	return false
}

func (r *River) WasAt(y, x int) bool {
	for i := range r.y {
		if r.y[i] == y && r.x[i] == x {
			return true
		}
	}
	return false
}

func (r *River) Move(y, x int) {
	r.pathStack = append(r.pathStack, len(r.y))
	r.y = append(r.y, y)
	r.x = append(r.x, x)
}

func (r *River) GoBack() {
	r.pathStack = r.pathStack[:len(r.pathStack)-1]
}

func (r *River) SetColor(c color.Color, grid *Grid) {
	for i := range r.y {
		grid[r.y[i]][r.x[i]].SetColor(c)
	}
}

type City struct {
	CenterY, CenterX int
	Y, X             []int
	Size             int
}

func NewCity(y, x int) *City {
	return &City{
		CenterY: y,
		CenterX: x,
		Y:       []int{y},
		X:       []int{x},
		Size:    0,
	}
}

func (c *City) Has(y, x int) bool {
	for i := range c.Y {
		if c.Y[i] == y && c.X[i] == x {
			return true
		}
	}
	return false
}

func (c *City) AddSquare(y, x int) {
	if c.Has(y, x) {
		return
	}
	c.Y = append(c.Y, y)
	c.X = append(c.X, x)
}

type Country struct {
	Cities           []*City
	Y, X             []int
	BorderY, BorderX []int
	Color            color.Color
	CG               *CountryGroup
}

func NewCountry(city *City, cg *CountryGroup, color color.Color) *Country {
	country := &Country{
		Cities:  []*City{city},
		Y:       city.Y,
		X:       city.X,
		BorderY: city.Y,
		BorderX: city.X,
		Color:   color,
		CG:      cg,
	}
	country.SharpenBorder()
	return country
}

func (c *Country) Surface() int {
	return len(c.Y)
}

func (c *Country) HasInBorder(y, x int) bool {
	for i := range c.BorderY {
		if c.BorderY[i] == y && c.BorderX[i] == x {
			return true
		}
	}
	return false
}

func (c *Country) Take(y, x int) {
	c.Y = append(c.Y, y)
	c.X = append(c.X, x)
	c.BorderY = append(c.BorderY, y)
	c.BorderX = append(c.BorderX, x)
}

func (c *Country) TakeCity(city *City) {
	c.Cities = append(c.Cities, city)
	c.Y = append(c.Y, city.Y...)
	c.X = append(c.X, city.X...)
	c.BorderY = append(c.BorderY, city.Y...)
	c.BorderX = append(c.BorderX, city.X...)
}

func (c *Country) Leave(y, x int) {
	for i := range c.Y {
		if c.Y[i] == y && c.X[i] == x {
			if i < len(c.Y)-1 {
				c.Y = append(c.Y[:i], c.Y[i+1:]...)
				c.X = append(c.X[:i], c.X[i+1:]...)
			} else {
				c.Y = c.Y[:i]
				c.X = c.X[:i]
			}
			break
		}
	}
	for i := range c.BorderY {
		if c.BorderY[i] == y && c.BorderX[i] == x {
			if i < len(c.BorderY)-1 {
				c.BorderY = append(c.BorderY[:i], c.BorderY[i+1:]...)
				c.BorderX = append(c.BorderX[:i], c.BorderX[i+1:]...)
			} else {
				c.BorderY = c.BorderY[:i]
				c.BorderX = c.BorderX[:i]
			}
			break
		}
	}
}

func (c *Country) ClosestCity(y, x int) *City {
	dist := -1
	var cc *City
	for _, city := range c.Cities {
		if dist == -1 || (city.CenterY-y)*(city.CenterY-y)+(city.CenterX-x)*(city.CenterX-x) < dist {
			cc = city
		}
	}
	return cc
}

func (c *Country) SharpenBorder() {
	ic := c.CG.Index(c)
	if ic == -1 {
		return
	}
	for i := 0; i < len(c.BorderY); {
		y, x := c.BorderY[i], c.BorderX[i]
		keep := false
		for _, dir := range DIR_NEXT {
			oy, ox := Inside(y+dir[0], x+dir[1])
			keep = keep || c.CG.Grid[oy][ox].CountryIndex != ic
		}
		if !keep {
			if i == len(c.BorderY)-1 {
				c.BorderY = c.BorderY[:i]
				c.BorderX = c.BorderX[:i]
			} else {
				c.BorderY = append(c.BorderY[:i], c.BorderY[i+1:]...)
				c.BorderX = append(c.BorderX[:i], c.BorderX[i+1:]...)
			}
		} else {
			i++
		}
	}
}

func (c *Country) Center() (centerY, centerX int) {
	for i := range c.Y {
		centerY += c.Y[i]
		centerX += c.X[i]
	}
	centerY /= len(c.Y)
	centerX /= len(c.X)
	return
}

type CountryGroup struct {
	countries []*Country
	Grid      *Grid
}

func NewCountryGroup(grid *Grid) *CountryGroup {
	return &CountryGroup{
		countries: []*Country{},
		Grid:      grid,
	}
}

func (cg *CountryGroup) AddCountry(country *Country) {
	cg.countries = append(cg.countries, country)
}

func (cg *CountryGroup) Index(country *Country) int {
	for i := range cg.countries {
		if cg.countries[i] == country {
			return i
		}
	}
	return -1
}

func (cg *CountryGroup) CountryCount() int {
	return len(cg.countries)
}

func (cg *CountryGroup) Get(i int) *Country {
	return cg.countries[i]
}

func (cg *CountryGroup) Surface() int {
	var s int
	for _, c := range cg.countries {
		s += c.Surface()
	}
	return s
}

func (cg *CountryGroup) HasInBorders(y, x int) bool {
	for _, c := range cg.countries {
		if c.HasInBorder(y, x) {
			return true
		}
	}
	return false
}

func AddFeaturesToTerrain(terrain [GRID_HEIGHT][GRID_WIDTH]int) *Grid {
	// inner model
	grid := &Grid{}
	for y := range grid {
		for x := range grid[y] {
			grid[y][x] = &SquareTerrain{
				Val:     terrain[y][x],
				Feature: FEATURE_NONE,
			}
		}
	}

	// delete isolated
	done := false
	for !done {
		done = true
		for y := range grid {
			for x := range grid[y] {
				surroundings := 0
				for _, dir := range DIRECTIONS {
					nhbY, nhbX := Inside(y+dir[0], x+dir[1])
					surroundings += grid[nhbY][nhbX].Val
				}
				if surroundings*grid[y][x].Val < 0 {
					grid[y][x].Val *= -1
					done = false
				}
			}
		}
	}

	// terrain: land and sea
	var nLand, nSea int
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
	println("Land:", nLand, "Sea:", nSea, "Land%:", nLand*100/SURFACE)

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
	rivers = append(rivers, NewRiver(grid))
	var riverSurface int
	for riverSurface < nLand*RIVER_PCT/100 {
		river := rivers[len(rivers)-1]
		// for each river not at sea yet, decide where to go
		highDir, highLevel := -1, river.Level
		end := false
		for _, dir := range rand.Perm(len(DIR_NEXT)) {
			nhbY, nhbX := Inside(river.Y()+DIR_NEXT[dir][0], river.X()+DIR_NEXT[dir][1])
			if river.WasAt(nhbY, nhbX) {
				continue
			}
			tight := false
			for _, diro := range DIR_NEXT {
				onY, onX := Inside(nhbY+diro[0], nhbX+diro[1])
				if (onY != river.Y() || onX != river.X()) && river.IsAt(onY, onX) {
					tight = true
					break
				}
			}
			if tight {
				continue
			}
			if grid[nhbY][nhbX].Feature == FEATURE_RIVER || grid[nhbY][nhbX].Terrain == TERRAIN_SEA {
				end = true
				break
			}
			if grid[nhbY][nhbX].Val >= highLevel {
				highDir = dir
				highLevel = grid[nhbY][nhbX].Val
			}
		}

		// act
		if end {
			// end: skip to next river
			rivers = append(rivers, NewRiver(grid))
			riverSurface += river.Len()
		} else if highDir == -1 {
			// go back
			if river.Len() > 1 {
				grid[river.Y()][river.X()].Feature = FEATURE_NONE
				oldY, oldX := river.Y(), river.X()
				river.GoBack()
				river.Level -= grid[oldY][oldX].Val - grid[river.Y()][river.X()].Val
			} else if river.Level >= 0 {
				river.Level--
			} else {
				rivers = append(rivers[:len(rivers)], NewRiver(grid))
			}
		} else {
			// move forward
			nhbY, nhbX := Inside(river.Y()+DIR_NEXT[highDir][0], river.X()+DIR_NEXT[highDir][1])
			grid[nhbY][nhbX].Feature = FEATURE_RIVER
			river.Level += grid[nhbY][nhbX].Val - grid[river.Y()][river.X()].Val
			river.Move(nhbY, nhbX)
		}
	}
	println(len(rivers), "rivers:", riverSurface)

	// cities
	var cities []*City
	for len(cities) < NB_CITIES {
		y, x := rand.Intn(GRID_HEIGHT), rand.Intn(GRID_WIDTH)
		if grid[y][x].Terrain != TERRAIN_LAND || grid[y][x].Feature != FEATURE_NONE {
			continue
		}

		city := NewCity(y, x)

		cityUpsizers := []func(st *SquareTerrain) bool{
			func(st *SquareTerrain) bool {
				return st.Terrain == TERRAIN_SEA
			},
			func(st *SquareTerrain) bool {
				return st.Feature == FEATURE_RIVER
			},
		}
		for _, dir := range DIRECTIONS {
			nhbY, nhbX := Inside(y+dir[0], x+dir[1])
			for i := range cityUpsizers {
				if cityUpsizers[i](grid[nhbY][nhbX]) {
					city.Size++
				}
			}
		}
		if rand.Intn(5) < 1 {
			city.Size++
		}

		if city.Size == 0 && rand.Intn(100) < 90 {
			continue
		}

		grid[y][x].Feature = FEATURE_CITY
		switch city.Size {
		case 2:
			for _, dir := range DIR_SQUARE {
				nhbY, nhbX := Inside(y+dir[0], x+dir[1])
				if grid[nhbY][nhbX].Terrain == TERRAIN_LAND && grid[nhbY][nhbX].Feature == FEATURE_NONE {
					grid[nhbY][nhbX].Feature = FEATURE_CITY
					v := uint8(grid[nhbY][nhbX].Val)
					grid[nhbY][nhbX].SetRGBA(v/2, v, 0, 255)
					city.AddSquare(nhbY, nhbX)
				}
			}
		case 3:
			for _, dir := range DIRECTIONS {
				nhbY, nhbX := Inside(y+dir[0], x+dir[1])
				if grid[nhbY][nhbX].Terrain == TERRAIN_LAND && grid[nhbY][nhbX].Feature == FEATURE_NONE {
					grid[nhbY][nhbX].Feature = FEATURE_CITY
					v := uint8(grid[nhbY][nhbX].Val)
					grid[nhbY][nhbX].SetRGBA(v/2, v, 0, 255)
					city.AddSquare(nhbY, nhbX)
				}
			}
		}
		cities = append(cities, city)
	}
	println(len(cities), "cities")

	// countries and borders
	for y := range grid {
		for x := range grid[y] {
			grid[y][x].CountryIndex = -1
		}
	}
	cg := NewCountryGroup(grid)
	for i := range rand.Perm(len(cities)) {
		if i >= NB_COUNTRIES {
			break
		}
		cg.AddCountry(NewCountry(cities[i], cg, HSVtoRGBA(i*240/NB_COUNTRIES, .5, .5)))
		for j := range cities[i].Y {
			grid[cities[i].Y[j]][cities[i].X[j]].CountryIndex = cg.CountryCount() - 1
		}
	}
	for done := false; !done; {
		done = true
		print("\r", cg.CountryCount(), " countries: ", 100*cg.Surface()/nLand)
	countriesLoop:
		for ic := 0; ic < cg.CountryCount(); ic++ {
			country := cg.Get(ic)
			country.SharpenBorder()
			for _, ib := range rand.Perm(len(country.BorderY)) {
				y, x := country.BorderY[ib], country.BorderX[ib]
				for _, dir := range DIR_NEXT {
					nhbY, nhbX := Inside(y+dir[0], x+dir[1])
					if grid[nhbY][nhbX].Terrain == TERRAIN_SEA || grid[nhbY][nhbX].Feature == FEATURE_RIVER {
						continue
					}
					if grid[nhbY][nhbX].CountryIndex == -1 {
						// take new square
						if grid[nhbY][nhbX].Feature == FEATURE_CITY {
							for _, city := range cities {
								if city.Has(nhbY, nhbX) {
									country.TakeCity(city)
									for j := range city.Y {
										grid[city.Y[j]][city.X[j]].CountryIndex = ic
									}
									break
								}
							}
						} else {
							country.Take(nhbY, nhbX)
							grid[nhbY][nhbX].CountryIndex = ic
						}
						done = false
						continue countriesLoop
					}
				}
			}
		}
	}
	println()
	for i := 0; i < cg.CountryCount(); i++ {
		country := cg.Get(i)
		for j := range country.BorderY {
			y, x := country.BorderY[j], country.BorderX[j]
			trace := true
			for _, dir := range DIR_NEXT {
				yo, xo := Inside(y+dir[0], x+dir[1])
				if grid[yo][xo].Terrain == TERRAIN_SEA || grid[yo][xo].Feature == FEATURE_RIVER {
					trace = false
					break
				}
				if grid[yo][xo].CountryIndex != i && grid[yo][xo].Feature == FEATURE_COUNTRY_BORDER {
					trace = false
					break
				}
			}
			if trace {
				grid[y][x].Feature = FEATURE_COUNTRY_BORDER
			}
		}
	}

	// map borders
	if !CONNECT_Y {
		for x := 0; x < GRID_WIDTH; x++ {
			grid[0][x].Terrain = TERRAIN_MAP_BORDER
			grid[GRID_HEIGHT-1][x].Terrain = TERRAIN_MAP_BORDER
			if x%MAGIC < MAGIC/2 {
				grid[0][x].SetRGBA(45, 45, 45, 255)
				grid[GRID_HEIGHT-1][x].SetRGBA(150, 150, 150, 255)
			} else {
				grid[0][x].SetRGBA(150, 150, 150, 255)
				grid[GRID_HEIGHT-1][x].SetRGBA(45, 45, 45, 255)
			}
		}
	}
	if !CONNECT_X {
		for y := 0; y < GRID_HEIGHT; y++ {
			grid[y][0].Terrain = TERRAIN_MAP_BORDER
			grid[y][GRID_WIDTH-1].Terrain = TERRAIN_MAP_BORDER
			if y%MAGIC < MAGIC/2 {
				grid[y][0].SetRGBA(150, 150, 150, 255)
				grid[y][GRID_WIDTH-1].SetRGBA(45, 45, 45, 255)
			} else {
				grid[y][0].SetRGBA(45, 45, 45, 255)
				grid[y][GRID_WIDTH-1].SetRGBA(150, 150, 150, 255)
			}
		}
	}

	// colors
	for y := range grid {
		for x := range grid[y] {
			switch v := uint8(grid[y][x].Val); grid[y][x].Terrain {
			case TERRAIN_LAND:
				grid[y][x].SetRGBA(v/2, v, 0, 255)
			case TERRAIN_MOUNTAIN:
				grid[y][x].SetRGBA(uint8(int(v)*224/255), uint8(int(v)*228/255), uint8(int(v)*170/255), 255)
			case TERRAIN_SEA:
				grid[y][x].SetRGBA(0, 0, uint8(255-int(v)*3/4), 255)
			}
		}
	}
	return grid
}
