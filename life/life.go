package life

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	rows = 20
	cols = 40

	aliveChar = '█'
	deadChar  = ' '

	// Every mutationInterval generations mutation happens
	mutationInterval = 50
	// Probability of nutation of every cell
	mutationRate = 0.02
)

type Grid [][]bool

func NewGrid(r, c int) Grid {
	g := make(Grid, r)
	for i := range g {
		g[i] = make([]bool, c)
	}
	return g
}

func (g Grid) getRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (g Grid) Randomize(p float64) {
	rnd := g.getRand()
	for i := range g {
		for j := range g[i] {
			g[i][j] = rnd.Float64() < p
		}
	}
}

func (g Grid) CountNeighbors(i, j int) int {
	r := len(g)
	c := len(g[0])
	count := 0
	for di := -1; di <= 1; di++ {
		for dj := -1; dj <= 1; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ni := (i + di + r) % r
			nj := (j + dj + c) % c
			if g[ni][nj] {
				count++
			}
		}
	}
	return count
}

func (g Grid) Next() Grid {
	r := len(g)
	c := len(g[0])
	next := NewGrid(r, c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			n := g.CountNeighbors(i, j)
			if g[i][j] {
				next[i][j] = (n == 2 || n == 3)
			} else {
				next[i][j] = (n == 3)
			}
		}
	}
	return next
}

// Mutate cell so as to avoid stable state and add a little chaos
func (g Grid) Mutate(p float64) {
	for i := range g {
		for j := range g[i] {
			if rand.Float64() < p {
				g[i][j] = !g[i][j] // invert state
			}
		}
	}
}

func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

func (g Grid) Print() {
	for i := range g {
		for j := range g[i] {
			if g[i][j] {
				fmt.Printf("%c", aliveChar)
			} else {
				fmt.Printf("%c", deadChar)
			}
		}
		fmt.Println()
	}
}

func Run() {
	g := NewGrid(rows, cols)
	g.Randomize(0.25)
	tick := 150 * time.Millisecond

	for gen := 0; ; gen++ {
		ClearScreen()
		fmt.Printf("Conway's Game of Life — generation %d\n\n", gen)
		g.Print()

		// Mutate every mutationInterval generations
		if gen > 0 && gen%mutationInterval == 0 {
			fmt.Println("\nMutation! (add chaos)")
			g.Mutate(mutationRate)
			time.Sleep(500 * time.Millisecond)
		}

		time.Sleep(tick)
		g = g.Next()
	}
}
