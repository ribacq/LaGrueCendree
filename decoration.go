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
				st.DecorateRiver(connections)
			case FEATURE_CITY:
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

func (st *SquareTerrain) DecorateRiver(connections [len(DIR_NEXT)]bool) {
	var forbidden [SQUARE_HEIGHT][SQUARE_WIDTH]bool
	for i, dir := range DIR_NEXT {
		if !connections[i] {
			if dir[0] == 0 {
				for sy := 0; sy < SQUARE_HEIGHT; sy++ {
					forbidden[sy][(SQUARE_WIDTH-dir[1])%(SQUARE_WIDTH+1)] = true
					forbidden[sy][(SQUARE_WIDTH-2*dir[1])%(SQUARE_WIDTH+1)] = true
				}
			} else if dir[1] == 0 {
				for sx := 0; sx < SQUARE_WIDTH; sx++ {
					forbidden[(SQUARE_HEIGHT-dir[0])%(SQUARE_HEIGHT+1)][sx] = true
					forbidden[(SQUARE_HEIGHT-2*dir[0])%(SQUARE_HEIGHT+1)][sx] = true
				}
			}
		}
	}
	for sy := range st.Colors {
		for sx := range st.Colors[sy] {
			if forbidden[sy][sx] {
				continue
			}
			if rand.Intn(SQUARE_HEIGHT*SQUARE_WIDTH/3) < 1 {
				r, g, b, a := st.Colors[sy][sx].RGBA()
				st.Colors[sy][sx] = color.RGBA{
					R: uint8(r),
					G: uint8(g),
					B: uint8((int(b) + 255) / 2),
					A: uint8(a),
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
