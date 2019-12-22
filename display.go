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
	TERRAIN_CITY
	TERRAIN_MAP_BORDER
	TERRAIN_COUNTRY_BORDER
)

var (
	RIVER_PCT    int = 8
	NB_CITIES    int = MAGIC
	NB_COUNTRIES int = MAGIC / 3
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

func HSVtoRGB(H int, S, V float64) (out Color) {
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
	out = Color{
		R: int((r + m) * 255),
		G: int((g + m) * 255),
		B: int((b + m) * 255),
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

type SquareTerrain struct {
	Val     int
	Terrain int
	Color   Color
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
	grid[y][x].Terrain = TERRAIN_RIVER
	return &River{
		y:         []int{y},
		x:         []int{x},
		pathStack: []int{0},
		Level:     grid[y][x].Val,
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

func (r *River) SetColor(color Color, grid *Grid) {
	for i := range r.y {
		grid[r.y[i]][r.x[i]].Color = color
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
	Color            Color
}

func NewCountry(city *City, color Color) *Country {
	country := &Country{
		Cities: []*City{city},
		Y:      city.Y,
		X:      city.X,
		Color:  color,
	}
	country.SharpenBorder()
	return country
}

func (c *Country) Surface() int {
	return len(c.Y)
}

func (c *Country) Has(y, x int) bool {
	for i := range c.Y {
		if c.Y[i] == y && c.X[i] == x {
			return true
		}
	}
	return false
}

func (c *Country) Take(y, x int) {
	c.Y = append(c.Y, y)
	c.X = append(c.X, x)
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
			return
		}
	}
}

func (c *Country) TakeCity(city *City) {
	c.Cities = append(c.Cities, city)
	c.Y = append(c.Y, city.Y...)
	c.X = append(c.X, city.X...)
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

func (c *Country) GetMaxDistSquared() int {
	cFirst := c.Cities[0]
	cLast := c.Cities[len(c.Cities)-1]
	mds := (cFirst.CenterY-cLast.CenterY)*(cFirst.CenterY-cLast.CenterY) + (cFirst.CenterX-cLast.CenterX)*(cFirst.CenterX-cLast.CenterX)
	if mds < MAGIC {
		return 0
	}
	return mds
}

func (c *Country) SharpenBorder() int {
	for i := range c.BorderY {
		for _, dir := range DIR_NEXT {
			nhbY, nhbX := Inside(c.BorderY[i]+dir[0], c.BorderX[i]+dir[1])
			// TODO: (*Country).SharpenBorder()
		}
	}
}

type CountryGroup struct {
	countries []*Country
}

func NewCountryGroup() *CountryGroup {
	return &CountryGroup{
		countries: []*Country{},
	}
}

func (cg *CountryGroup) AddCountry(country *Country) {
	cg.countries = append(cg.countries, country)
}

func (cg *CountryGroup) CountryCount() int {
	return len(cg.countries)
}

func (cg *CountryGroup) Get(i int) *Country {
	return cg.countries[i]
}

func (cg *CountryGroup) GetCountryAt(y, x int) (c *Country, ok bool) {
	for _, c := range cg.countries {
		if c.Has(y, x) {
			return c, true
		}
	}
	return c, false
}

func (cg *CountryGroup) Surface() int {
	var s int
	for _, c := range cg.countries {
		s += c.Surface()
	}
	return s
}

func (cg *CountryGroup) Has(y, x int) bool {
	for _, c := range cg.countries {
		if c.Has(y, x) {
			return true
		}
	}
	return false
}

type Grid [GRID_HEIGHT][GRID_WIDTH]SquareTerrain

func display(squares [GRID_HEIGHT][GRID_WIDTH]int) (nLand, nSea int) {
	// inner model
	grid := &Grid{}
	for y := range grid {
		for x := range grid[y] {
			grid[y][x].Val = squares[y][x]
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
			if grid[nhbY][nhbX].Terrain == TERRAIN_RIVER || grid[nhbY][nhbX].Terrain == TERRAIN_SEA {
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
			//println(len(rivers), "y:", river.Y(), "x:", river.X(), "len:", river.Len(), "level:", river.Level)
			rivers = append(rivers, NewRiver(grid))
			riverSurface += river.Len()
		} else if highDir == -1 {
			// go back
			if river.Len() > 1 {
				grid[river.Y()][river.X()].Terrain = TERRAIN_LAND
				oldY, oldX := river.Y(), river.X()
				river.GoBack()
				river.Level -= grid[oldY][oldX].Val - grid[river.Y()][river.X()].Val
			} else {
				river.Level--
			}
		} else {
			// move forward
			nhbY, nhbX := Inside(river.Y()+DIR_NEXT[highDir][0], river.X()+DIR_NEXT[highDir][1])
			grid[nhbY][nhbX].Terrain = TERRAIN_RIVER
			river.Level += grid[nhbY][nhbX].Val - grid[river.Y()][river.X()].Val
			river.Move(nhbY, nhbX)
		}
	}
	println(len(rivers), "rivers: ", riverSurface)

	// cities
	var cities []*City
	for len(cities) < NB_CITIES {
		y, x := rand.Intn(GRID_HEIGHT), rand.Intn(GRID_WIDTH)
		if grid[y][x].Terrain != TERRAIN_LAND {
			continue
		}

		city := NewCity(y, x)

		cityUpsizers := map[int]bool{
			TERRAIN_SEA:   false,
			TERRAIN_RIVER: false,
		}
		for _, dir := range DIRECTIONS {
			nhbY, nhbX := Inside(y+dir[0], x+dir[1])
			for terrain := range cityUpsizers {
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
		if rand.Intn(5) < 1 {
			city.Size++
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
					city.AddSquare(nhbY, nhbX)
				}
			}
		case 3:
			grid[y][x].Color = Color{255, 255, 255}
			for _, dir := range DIRECTIONS {
				nhbY, nhbX := Inside(y+dir[0], x+dir[1])
				if grid[nhbY][nhbX].Terrain == TERRAIN_LAND {
					grid[nhbY][nhbX].Terrain = TERRAIN_CITY
					grid[nhbY][nhbX].Color = Color{232, 13, 17}
					city.AddSquare(nhbY, nhbX)
				}
			}
		}
		cities = append(cities, city)
	}
	println(len(cities), "cities")

	// countries and borders
	cg := NewCountryGroup()
	for i := range rand.Perm(len(cities)) {
		if i >= NB_COUNTRIES {
			break
		}
		cg.AddCountry(NewCountry(cities[i], Color{
			R: rand.Intn(256),
			G: rand.Intn(256),
			B: rand.Intn(256),
		}))
	}
	done = false
	for !done {
		done = true
		print("\r", cg.CountryCount(), " countries: ", cg.Surface()*100/nLand)
		for i := 0; i < cg.CountryCount(); i++ {
			country := cg.Get(i)
			for i := range country.Y {
				y, x := country.Y[i], country.X[i]
				cc := country.ClosestCity(y, x)
				distCCsquared := (cc.CenterY-y)*(cc.CenterY-y) + (cc.CenterX-x)*(cc.CenterX-x)
				mds := country.GetMaxDistSquared()
				if mds != 0 && distCCsquared > mds {
					continue
				}
				for _, dir := range DIR_NEXT {
					nhbY, nhbX := Inside(y+dir[0], x+dir[1])
					if grid[nhbY][nhbX].Terrain == TERRAIN_SEA {
						continue
					}
					if !cg.Has(nhbY, nhbX) {
						// take new square
						if grid[nhbY][nhbX].Terrain == TERRAIN_CITY {
							for _, city := range cities {
								if city.Has(nhbY, nhbX) {
									country.TakeCity(city)
									break
								}
							}
						} else {
							country.Take(nhbY, nhbX)
						}
						done = false
					} else if oc, _ := cg.GetCountryAt(nhbY, nhbX); oc != country && oc.GetMaxDistSquared() > country.GetMaxDistSquared() {
						oc.Leave(nhbY, nhbX)
						country.Take(nhbY, nhbX)
						done = false
					}
				}
			}
		}
	}
	println()

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
			case TERRAIN_RIVER:
				grid[y][x].Color = Color{
					R: v / 4,
					G: v / 2,
					B: v,
				}
			case TERRAIN_COUNTRY_BORDER:
				grid[y][x].Color = Color{255, 0, 0}
			case TERRAIN_SEA:
				grid[y][x].Color = Color{
					R: 0,
					G: 0,
					B: 255 - v*3/4,
				}
			}
			switch grid[y][x].Terrain {
			case TERRAIN_LAND, TERRAIN_MOUNTAIN, TERRAIN_CITY:
				country, ok := cg.GetCountryAt(y, x)
				if ok {
					grid[y][x].Color = Color{
						R: (grid[y][x].Color.R + country.Color.R) / 2,
						G: (grid[y][x].Color.G + country.Color.G) / 2,
						B: (grid[y][x].Color.B + country.Color.B) / 2,
					}
				} else {
					grid[y][x].Color = Color{0, 0, 0}
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
