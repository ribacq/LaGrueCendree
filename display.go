package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
)

func PrintPPM(grid *Grid) {
	fmt.Println("P3")
	fmt.Println(GRID_WIDTH, GRID_HEIGHT, 255)
	for y := range grid {
		for x := range grid[y] {
			for sy := range grid[y][x].Colors {
				for sx := range grid[y][x].Colors[sy] {
					r, g, b, _ := grid[y][x].Colors[sy][sx].RGBA()
					fmt.Print(r, g, b, " ")
				}
			}
		}
		fmt.Println()
	}
}

func PrintPNG(grid *Grid) {
	img := image.NewRGBA(image.Rect(0, 0, len(grid[0])*SQUARE_WIDTH, len(grid)*SQUARE_HEIGHT))
	for y := range grid {
		for x := range grid[y] {
			for sy := range grid[y][x].Colors {
				for sx := range grid[y][x].Colors[sy] {
					img.Set(x*SQUARE_WIDTH+sx, y*SQUARE_HEIGHT+sy, grid[y][x].Colors[sy][sx])
				}
			}
		}
	}

	err := png.Encode(os.Stdout, img)
	if err != nil {
		panic(err)
	}
}
