package main

import (
	"fmt"
	"image/color"
	"math/rand"
)

func (grid *Grid) DecorateFeatures() {
	for y := range grid {
		for x := range grid[y] {
			switch st := grid[y][x]; st.Feature {
			case FEATURE_RIVER:
				var connections [len(DIR_NEXT)]bool
				for i, dir := range DIR_NEXT {
					ny, nx := Inside(y+dir[0], x+dir[1])
					if grid[ny][nx].Terrain == TERRAIN_SEA || grid[ny][nx].Feature == FEATURE_RIVER {
						connections[i] = true
					}
				}
				st.DrawRiver(connections)
			case FEATURE_CITY:
				st.DrawHouse()
			case FEATURE_COUNTRY_BORDER:
				st.DrawCountryBorder()
			}
		}
	}
}

func (st *SquareTerrain) Draw(shape string, colors map[byte]color.Color) error {
	if len(shape) != SQUARE_HEIGHT*SQUARE_WIDTH {
		return fmt.Errorf("invalid shape size %v, expected %v", len(shape), SQUARE_HEIGHT*SQUARE_WIDTH)
	}

	for y := range st.Colors {
		for x := range st.Colors[y] {
			if c := colors[shape[y*SQUARE_WIDTH+x]]; c != color.Transparent {
				st.Colors[y][x] = c
			}
		}
	}
	return nil
}

func (st *SquareTerrain) DrawRiver(connections [len(DIR_NEXT)]bool) {
	var riverY, riverX []int
	for i, dir := range DIR_NEXT {
		if connections[i] {
			if dir[0] == 0 {
				for sy := 2; sy < SQUARE_WIDTH-2; sy++ {
					riverY = append(riverY, sy, sy)
					riverX = append(riverX, (SQUARE_WIDTH-dir[1])%(SQUARE_WIDTH+1), (SQUARE_WIDTH-2*dir[1])%(SQUARE_WIDTH+1))
				}
			} else if dir[1] == 0 {
				for sx := 2; sx < SQUARE_HEIGHT-2; sx++ {
					riverY = append(riverY, (SQUARE_HEIGHT-dir[0])%(SQUARE_HEIGHT+1), (SQUARE_HEIGHT-2*dir[0])%(SQUARE_HEIGHT+1))
					riverX = append(riverX, sx, sx)
				}
			}
		}
	}
	for sy := 2; sy < len(st.Colors)-2; sy++ {
		for sx := 2; sx < len(st.Colors[sy])-2; sx++ {
			if rand.Intn(SQUARE_HEIGHT*SQUARE_WIDTH/3) < 1 {
				st.Colors[sy][sx] = color.RGBA{
					R: uint8(st.Val / 4),
					G: uint8(st.Val / 2),
					B: uint8((st.Val + 255) / 2),
					A: 255,
				}
			} else {
				st.Colors[sy][sx] = color.RGBA{
					R: uint8(st.Val / 4),
					G: uint8(st.Val / 2),
					B: uint8(st.Val),
					A: 255,
				}
			}
		}
	}
	for i := range riverY {
		st.Colors[riverY[i]][riverX[i]] = color.RGBA{
			R: uint8(st.Val / 4),
			G: uint8(st.Val / 2),
			B: uint8(st.Val),
			A: 255,
		}
	}
}

func (st *SquareTerrain) DrawHouse() {
	colors := map[byte]color.Color{
		't': color.RGBA{144, 71, 17, 255},
		's': color.RGBA{83, 42, 8, 255},
		'w': color.RGBA{203, 101, 27, 255},
		'd': color.RGBA{156, 167, 174, 255},
		'n': color.RGBA{212, 220, 224, 255},
		'.': color.Transparent,
	}

	shape := "..tsww...tswwww.tswwwwwwttttttttssssssstwwdwnwstwwdwwwsttttttttt"

	err := st.Draw(shape, colors)
	if err != nil {
		panic(err)
	}
}

func (st *SquareTerrain) DrawCountryBorder() {
	for sy := 0; sy < 2; sy++ {
		for sx := 0; sx < 2; sx++ {
			st.Colors[SQUARE_HEIGHT/2-1+sy][SQUARE_WIDTH/2-1+sx] = color.RGBA{255, 0, 0, 255}
		}
	}
}
