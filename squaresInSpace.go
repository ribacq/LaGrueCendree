// See LICENSE file for legal info
// Quentin RIBAC, december 2019
package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	GRID_WIDTH  int = 640
	GRID_HEIGHT int = 360
	FRAMES      int = 50
	DIAGONAL    int = GRID_WIDTH + GRID_HEIGHT
	MAX_DIST    int = 32
	MAX_VAL     int = MAX_DIST
	START_VAL   int = MAX_VAL
	RENEW       int = MAX_DIST
	EMPTY       int = 0
)

var (
	DIRECTIONS [4][2]int = [4][2]int{
		{0, 1},
		{0, -1},
		{1, 0},
		{-1, 0},
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
	if y < 0 {
		for y < 0 {
			y += GRID_HEIGHT
		}
	}
	if y >= GRID_HEIGHT {
		for y >= GRID_HEIGHT {
			y -= GRID_HEIGHT
		}
	}
	if x < 0 {
		for x < 0 {
			x += GRID_WIDTH
		}
	}
	if x >= GRID_WIDTH {
		for x >= GRID_WIDTH {
			x -= GRID_WIDTH
		}
	}
	return y, x
}

func main() {
	var squares [GRID_HEIGHT][GRID_WIDTH]int
	rand.Seed(time.Now().UnixNano())

	println(GRID_WIDTH, "x", GRID_HEIGHT)

	// main loop
	for frame := 0; frame < FRAMES; frame++ {
		// loop on squares
		for y := range squares {
			print("              \r", frame, " ", y)
			for x := range squares[y] {
				go func(y, x int) {
					// life creation
					if rand.Intn(RENEW) < 1 {
						if squares[y][x] == EMPTY {
							squares[y][x] = START_VAL
						} else if squares[y][x] < EMPTY {
							squares[y][x]++
						}
					}

					// next position
					var nextY, nextX int

					// skip empty
					if squares[y][x] < EMPTY {
						squares[y][x]++
					}

					if squares[y][x] <= EMPTY {
						return
					}

					// move to other
					dy, dx := 0, 0
					dist := squares[y][x]
					for yo := y - dist; yo <= y+dist; yo++ {
						for xo := x - dist + Abs(yo-y); xo <= x+dist-Abs(yo-y); xo++ {
							// skip empty squares
							inyo, inxo := Inside(yo, xo)
							if squares[inyo][inxo] <= EMPTY {
								continue
							}

							// count force
							dy += (yo - y)
							dx += (xo - x)
						}
					}
					dx = Sign(dx)
					dy = Sign(dy)
					nextY, nextX = Inside(y+dy, x+dx)
					if (nextY != y || nextX != x) && squares[nextY][nextX] <= EMPTY {
						squares[nextY][nextX] = squares[y][x]
						squares[y][x] = -squares[y][x] / 4
					}

					// stay still: slowly die
					squares[y][x]--
				}(y, x)
			}
		}
	}

	// display
	fmt.Println("P3")
	fmt.Println(GRID_WIDTH, GRID_HEIGHT, MAX_VAL-1)
	for y := range squares {
		for x := range squares[y] {
			if squares[y][x] == EMPTY {
				fmt.Print(0, 0, 0, " ")
			} else if squares[y][x] < EMPTY {
				fmt.Print(0, 0, -squares[y][x], " ")
			} else {
				fmt.Print(0, MAX_VAL-squares[y][x], 0, " ")
			}
		}
		fmt.Println()
	}

	// end
	println()
}
