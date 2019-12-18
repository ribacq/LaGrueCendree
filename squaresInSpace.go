// See LICENSE file for legal info
// Quentin RIBAC, december 2019
package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// make your choice here
const (
	GRID_WIDTH  int  = 300
	GRID_HEIGHT int  = 150
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
	DIRECTIONS [8][2]int = [8][2]int{
		{0, 1}, {0, -1}, {1, 0}, {-1, 0},
		{0, 2}, {0, -2}, {2, 0}, {-2, 0},
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

func display(squares [GRID_HEIGHT][GRID_WIDTH]int, cities bool) (nP, nN int) {
	// delete isolated
	for _, y := range rand.Perm(GRID_HEIGHT) {
		for _, x := range rand.Perm(GRID_WIDTH) {
			surroundings := 0
			for dir := range DIRECTIONS {
				nhbY, nhbX := Inside(y+DIRECTIONS[dir][0], x+DIRECTIONS[dir][1])
				surroundings += squares[nhbY][nhbX]
			}
			if surroundings*squares[y][x] < 0 {
				squares[y][x] *= -1
			}
		}
	}

	// display
	fmt.Println("P3")
	fmt.Println(GRID_WIDTH, GRID_HEIGHT, MAX_VAL-1)
	for y := 0; y < GRID_HEIGHT; y++ {
		for x := 0; x < GRID_WIDTH; x++ {
			val := squares[y][x]
			for dir := range DIRECTIONS {
				nhbY, nhbX := Inside(y+DIRECTIONS[dir][0], x+DIRECTIONS[dir][1])
				val += squares[nhbY][nhbX]
			}
			val /= 1 + len(DIRECTIONS)
			if val <= -MAX_VAL {
				val = 1 - MAX_VAL
			}
			if val >= MAX_VAL {
				val = MAX_VAL - 1
			}
			if val <= 0 {
				fmt.Print(0, 0, -val/2, " ")
				nN++
			} else {
				if cities && squares[y][x] <= 0 && rand.Intn(4) < 1 {
					fmt.Print(MAX_VAL/2, 0, 0, " ")
				} else {
					fmt.Print(rand.Intn(val)/2, val, 0, " ")
				}
				nP++
			}
		}
		fmt.Println()
	}

	return nP, nN
}

func main() {
	var squares [GRID_HEIGHT][GRID_WIDTH]int
	rand.Seed(time.Now().UnixNano())

	println(GRID_WIDTH, "x", GRID_HEIGHT, "f", FRAMES)

	// spawn
	for i := 0; i < SPAWN_POWER; i++ {
		newY, newX := rand.Intn(GRID_HEIGHT), rand.Intn(GRID_WIDTH)
		if squares[newY][newX] == 0 {
			squares[newY][newX] = MAX_VAL
		}
	}

	// main loop on frames
	var wg sync.WaitGroup
	for frame := 1; frame <= FRAMES; frame++ {
		// loop on squares
		sy := 0
		for _, y := range rand.Perm(GRID_HEIGHT) {
			sy++
			print("              \r", frame, " ", sy)
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
					if squares[y][x] < 0 && squares[y][x] > -MAX_VAL {
						dy /= -squares[y][x]
						dx /= -squares[y][x]
					}
					if dist > 0 {
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

					squares[y][x]--
				}(y, x)
			}
		}
		wg.Wait()

		// display
		if GIF {
			display(squares, false)
		}
	}

	// end
	nP, nN := display(squares, true)
	println("\nP:", nP, "N:", nN, "P%:", 100*nP/SURFACE)
}
