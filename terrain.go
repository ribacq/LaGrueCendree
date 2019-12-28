package main

import (
	"math"
	"math/rand"
	"sync"
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
	DIRECTIONS [20][2]int = [20][2]int{
		{0, 1}, {0, -1}, {1, 0}, {-1, 0},
		{-1, -1}, {-1, 1}, {1, -1}, {1, 1},
		{0, 2}, {0, -2}, {2, 0}, {-2, 0},
		{-2, -1}, {-2, 1}, {-1, 2}, {1, 2},
		{2, 1}, {2, -1}, {1, -2}, {-1, -2},
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

func GenerateTerrain() [GRID_HEIGHT][GRID_WIDTH]int {
	var squares [GRID_HEIGHT][GRID_WIDTH]int
	println(GRID_WIDTH, "x", GRID_HEIGHT, "f", FRAMES)

	// spawn
	newY, newX := rand.Intn(GRID_HEIGHT), rand.Intn(GRID_WIDTH)
	for i := 0; i < SPAWN_POWER; i++ {
		if squares[newY][newX] == 0 {
			squares[newY][newX] = MAX_VAL
		}
		newY, newX = Inside(newY+rand.Intn(MAX_VAL), newX+rand.Intn(MAX_VAL))
	}

	// main loop on frames
	var wg sync.WaitGroup
	for frame := 1; frame <= FRAMES; frame++ {
		// loop on squares
		sy := 0
		for _, y := range rand.Perm(GRID_HEIGHT) {
			sy++
			print("\rframe ", frame, " (", sy, ")")
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
					if squares[y][x] < 0 {
						dy /= -squares[y][x]
						dx /= -squares[y][x]
					} else if dist > 0 {
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
					for _, dir := range rand.Perm(len(DIRECTIONS)) {
						nextY, nextX = Inside(y+DIRECTIONS[dir][0], x+DIRECTIONS[dir][1])
						if squares[nextY][nextX]*squares[y][x] <= 0 {
							squares[nextY][nextX], squares[y][x] = squares[y][x]+Sign(squares[y][x]), squares[nextY][nextX]-Sign(squares[y][x])
							return
						}
					}
				}(y, x)
			}
		}
		wg.Wait()
	}
	println()

	return squares
}

func GenerateTerrainQuick() [GRID_HEIGHT][GRID_WIDTH]int {
	var squares [GRID_HEIGHT][GRID_WIDTH]int
	for y := range squares {
		for x := range squares[y] {
			squares[y][x] = rand.Intn(len(DIRECTIONS)) - len(DIRECTIONS)*50/100
		}
	}
	for i, y := range rand.Perm(len(squares)) {
		print("\r", i)
		for _, x := range rand.Perm(len(squares[y])) {
			var s int
			for _, dir := range DIRECTIONS {
				yo, xo := Inside(y+dir[0], x+dir[1])
				s += Sign(squares[yo][xo])
			}
			squares[y][x] += s
			squares[y][x] /= 2
		}
	}
	println()

	return squares
}
