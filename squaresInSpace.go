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
	GRID_WIDTH  int = 1 + 256
	GRID_HEIGHT int = 1 + 144
	DIAGONAL    int = GRID_WIDTH + GRID_HEIGHT
	SURFACE     int = GRID_WIDTH * GRID_HEIGHT
)

var (
	MAX_DIST    int = DIAGONAL / 9
	FRAMES      int = MAX_DIST
	MAX_VAL     int = MAX_DIST
	SPAWNS      int = MAX_DIST * 9
	SPAWN_POWER int = MAX_DIST / 2
	SPAWN_DIST  int = MAX_DIST / 2
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

func main() {
	var squares [GRID_HEIGHT][GRID_WIDTH]int
	rand.Seed(time.Now().UnixNano())

	println(GRID_WIDTH, "x", GRID_HEIGHT, "f", FRAMES)

	// spawns creation
	spawns := make([]struct{ Y, X int }, SPAWNS)
	for i := range spawns {
		spawns[i].Y = rand.Intn(GRID_HEIGHT)
		spawns[i].X = rand.Intn(GRID_WIDTH)
	}

	// main loop on frames
	var wg sync.WaitGroup
	for frame := 0; frame < FRAMES; frame++ {

		// spawns action
		for _, sp := range spawns {
			for i := 0; i < SPAWN_POWER; i++ {
				offY, offX := rand.Intn(2*SPAWN_DIST)-SPAWN_DIST, rand.Intn(2*SPAWN_DIST)-SPAWN_DIST
				newY, newX := Inside(sp.Y+offY, sp.X+offX)
				if offY*offY+offX*offX < SPAWN_DIST*SPAWN_DIST && squares[newY][newX] == 0 {
					if rand.Intn(2) < 1 {
						squares[newY][newX] = MAX_VAL
					} else {
						squares[newY][newX] = -1
					}
				}
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
					if squares[y][x] <= 0 {
						return
					}

					if squares[y][x] <= -MAX_VAL {
						squares[y][x] = 1 - MAX_VAL
					}

					// move to other
					dy, dx := 0, 0
					dist := (squares[y][x] + MAX_VAL) % MAX_VAL
					for yo := y - dist; yo <= y+dist; yo++ {
						for xo := x - dist; xo <= x+dist; xo++ {
							// yo, xo inside
							inyo, inxo := Inside(yo, xo)

							// skip empty and far away
							if squares[inyo][inxo]*squares[y][x] <= 0 || (yo-y)*(yo-y)+(xo-x)*(xo-x) > dist*dist {
								continue
							}

							// count force
							dy += (inyo - y)
							dx += (inxo - x)
						}
					}
					if dist > 0 && squares[y][x] != 0 {
						dy /= dist
						dx /= dist
					}
					nextY, nextX := Inside(y+dy, x+dx)
					ry, rx := Abs(dy), Abs(dx)
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
					} else if ry == 0 && rx == 0 {
						dir := rand.Intn(len(DIRECTIONS))
						nextY, nextX = Inside(y+DIRECTIONS[dir][0], x+DIRECTIONS[dir][1])
						i := 0
						for i < len(DIRECTIONS) && squares[nextY][nextX]*squares[y][x] > 0 {
							dir = (dir + 1) % len(DIRECTIONS)
							nextY, nextX = Inside(y+DIRECTIONS[dir][0], x+DIRECTIONS[dir][1])
							i++
						}
						if i < len(DIRECTIONS) {
							squares[nextY][nextX], squares[y][x] = squares[y][x]-Sign(squares[y][x]), squares[nextY][nextX]-Sign(squares[y][x])
						}
					}
				}(y, x)
			}
		}
		wg.Wait()

		// display
		fmt.Println("P3")
		fmt.Println(GRID_WIDTH, GRID_HEIGHT, MAX_VAL-1)
		for y := range squares {
			for x := range squares[y] {
				val := squares[y][x]
				if y < GRID_HEIGHT-1 && x < GRID_WIDTH-1 {
					val = (squares[y][x] + squares[y+1][x] + squares[y][x+1] + squares[y+1][x+1]) / 4
				} else {
					//continue
				}
				if val <= 0 {
					fmt.Print(0, 0, -val, " ")
				} else if val > 0 {
					fmt.Print(rand.Intn(val), val, 0, " ")
				}
			}
			fmt.Println()
		}
	}

	// end
	println()
}
