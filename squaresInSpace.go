// See LICENSE file for legal info
// Quentin RIBAC, december 2019
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	GRID_WIDTH  int  = 1 + 256
	GRID_HEIGHT int  = 1 + 144
	CONNECT_Y   bool = false
	CONNECT_X   bool = true
	GIF         bool = false
)

var (
	DIAGONAL    int = GRID_WIDTH + GRID_HEIGHT
	SURFACE     int = GRID_WIDTH * GRID_HEIGHT
	MAGIC       int = DIAGONAL / 9
	FRAMES      int = MAGIC
	MAX_VAL     int = MAGIC
	SPAWN_POWER int = MAGIC
)

var (
	DIRECTIONS [8][2]int = [8][2]int{
		{0, 1},
		{0, -1},
		{1, 0},
		{-1, 0},
		{-1, -1}, {-1, 1}, {1, -1}, {1, 1},
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

type Spawn struct {
	Y, X  int
	Power int
}

func main() {
	var squares [GRID_HEIGHT][GRID_WIDTH]int
	rand.Seed(time.Now().UnixNano())

	println(GRID_WIDTH, "x", GRID_HEIGHT, "f", FRAMES)

	// main loop on frames
	var wg sync.WaitGroup
	for frame := 0; frame < FRAMES; frame++ {
		// spawns action
		for i := 0; i < SPAWN_POWER; i++ {
			newY, newX := rand.Intn(GRID_HEIGHT), rand.Intn(GRID_WIDTH)
			if squares[newY][newX] == 0 {
				squares[newY][newX] = MAX_VAL
			}
		}

		// loop on squares
		var seenY [GRID_HEIGHT]bool
		for range squares {
			y := rand.Intn(GRID_HEIGHT)
			for seenY[y] {
				y = rand.Intn(GRID_HEIGHT)
			}
			seenY[y] = true
			print("              \r", frame, " ", y)
			var seenX [GRID_WIDTH]bool
			for range squares[y] {
				x := rand.Intn(GRID_WIDTH)
				for seenX[x] {
					x = rand.Intn(GRID_WIDTH)
				}
				seenX[x] = true
				wg.Add(1)
				go func(y, x int) {
					defer wg.Done()

					// skip empty
					if squares[y][x] == 0 {
						return
					}

					if squares[y][x] <= -MAX_VAL {
						squares[y][x] = 1 - MAX_VAL
					}

					// move to other
					dy, dx := 0, 0
					dist := ((squares[y][x] + MAX_VAL) % MAX_VAL)
					for yo := y - dist; yo <= y+dist; yo++ {
						for xo := x - dist; xo <= x+dist; xo++ {
							// yo, xo inside
							inyo, inxo := Inside(yo, xo)

							// skip empty and far away
							if squares[inyo][inxo]*squares[y][x] <= 0 || (yo-y)*(yo-y)+(xo-x)*(xo-x) > dist*dist {
								continue
							}

							// count force
							if (CONNECT_Y || inyo == yo) && yo != y {
								dy += (yo - y)
							}
							if (CONNECT_X || inxo == xo) && xo != x {
								dx += (xo - x)
							}
						}
					}
					ry, rx := Abs(dy), Abs(dx)
					if dist > 0 && squares[y][x] != 0 {
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
					dir := rand.Intn(len(DIRECTIONS))
					nextY, nextX = Inside(y+DIRECTIONS[dir][0], x+DIRECTIONS[dir][1])
					i := 0
					for i < len(DIRECTIONS) {
						dir = (dir + 1) % len(DIRECTIONS)
						nextY, nextX = Inside(y+DIRECTIONS[dir][0], x+DIRECTIONS[dir][1])
						if squares[nextY][nextX]*squares[y][x] <= 0 {
							break
						}
						i++
					}
					if i < len(DIRECTIONS) {
						squares[nextY][nextX], squares[y][x] = squares[y][x]+Sign(squares[y][x]), squares[nextY][nextX]-Sign(squares[y][x])
						return
					}
				}(y, x)
			}
		}
		wg.Wait()

		// display
		if GIF {
			fmt.Println("P3")
			fmt.Println(GRID_WIDTH, GRID_HEIGHT, MAX_VAL-1)
			for y := range squares {
				for x := range squares[y] {
					val := squares[y][x]
					if y < GRID_HEIGHT-1 && x < GRID_WIDTH-1 {
						rpy := Abs(Sign(squares[y+1][x]))
						rpx := Abs(Sign(squares[y][x+1]))
						rpyx := Abs(Sign(squares[y+1][x+1]))
						val = (squares[y][x] + squares[y+1][x]*rpy + squares[y][x+1]*rpx + squares[y+1][x+1]*rpyx) / (1 + rpy + rpx + rpyx)
					}
					if val <= 0 {
						fmt.Print(0, 0, -val, " ")
					} else if val > 0 {
						fmt.Print(0, val, 0, " ")
					}
				}
				fmt.Println()
			}
		}
	}

	// delete isolated positives
	for y := range squares {
		for x := range squares[y] {
			if squares[y][x] > 0 {
				surroundings := 0
				for dir := range DIRECTIONS {
					nhbY, nhbX := Inside(y+DIRECTIONS[dir][0], x+DIRECTIONS[dir][1])
					surroundings += (squares[nhbY][nhbX])
				}
				if surroundings <= 0 {
					squares[y][x] *= -1
				}
			}
		}
	}

	// display
	fmt.Println("P3")
	fmt.Println(GRID_WIDTH, GRID_HEIGHT, MAX_VAL-1)
	for y := range squares {
		for x := range squares[y] {
			val := squares[y][x]
			if y < GRID_HEIGHT-1 && x < GRID_WIDTH-1 {
				rpy := Abs(Sign(squares[y+1][x]))
				rpx := Abs(Sign(squares[y][x+1]))
				rpyx := Abs(Sign(squares[y+1][x+1]))
				val = (squares[y][x] + squares[y+1][x]*rpy + squares[y][x+1]*rpx + squares[y+1][x+1]*rpyx) / 4
			}
			if val <= 0 {
				fmt.Print(0, 0, -val, " ")
			} else if val > 0 {
				fmt.Print(0, val, 0, " ")
			}
		}
		fmt.Println()
	}

	// end
	println()
}
