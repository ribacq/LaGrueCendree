package main

import (
	"fmt"
	"image/color"
	"math/rand"
)

func (grid *Grid) DecorateFeatures() {
	for y := range grid {
		for x := range grid[y] {
			switch st := grid[y][x]; st.Terrain {
			case TERRAIN_RIVER:
				st.DecorateRiver()
			case TERRAIN_CITY:
				st.DrawHouse()
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

func (st *SquareTerrain) DecorateRiver() {
	for sy := range st.Colors {
		for sx := range st.Colors[sy] {
			if rand.Intn(SQUARE_HEIGHT*SQUARE_WIDTH/3) < 1 {
				r, g, b, a := st.Colors[sy][sx].RGBA()
				st.Colors[sy][sx] = color.RGBA{
					R: uint8(r),
					G: uint8(g),
					B: uint8((int(b) + 255) / 2),
					A: uint8(a),
				}
			}
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
