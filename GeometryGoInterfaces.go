// draw.go
// Shapes and display implementation
// Zarak Khan

package main

import (
	"errors"
	"fmt"
	"math"
	"os"
)

// Basic color, point, and shape definitions

type Color string

// Map a readable color name to its RGB triplet.
var colorMap = map[Color][3]int{
	"red":    {255, 0, 0},
	"green":  {0, 255, 0},
	"blue":   {0, 0, 255},
	"yellow": {255, 255, 0},
	"orange": {255, 164, 0},
	"purple": {128, 0, 128},
	"brown":  {165, 42, 42},
	"black":  {0, 0, 0},
	"white":  {255, 255, 255},
}

type Point struct{ x, y int }

type Rectangle struct {
	ll, ur Point
	c      Color
}

type Triangle struct {
	pt0, pt1, pt2 Point
	c             Color
}

type Circle struct {
	center Point
	r      int
	c      Color
}

type Display struct {
	maxX, maxY int
	matrix     [][]Color
}

// Interfaces

type geometry interface {
	draw(scn screen) error
	printShape() string
}

type screen interface {
	initialize(x, y int)
	getMaxXY() (int, int)
	drawPixel(x, y int, c Color) error
	getPixel(x, y int) (Color, error)
	clearScreen()
	screenShot(f string) error
}

// Package-level helpers and errors

var (
	outOfBoundsErr  = errors.New("**Error: Attempt to draw a figure out of bounds of the screen.")
	colorUnknownErr = errors.New("**Error: Attempt to use an invalid color.")
)

// Return true if the color is not in the map.
func colorUnknown(c Color) bool {
	_, ok := colorMap[c]
	return !ok
}

// Return true if point p lies outside the screen.
func outOfBounds(p Point, scn screen) bool {
	mx, my := scn.getMaxXY()
	return p.x < 0 || p.x >= mx || p.y < 0 || p.y >= my
}

// Rectangle

func (r Rectangle) draw(scn screen) error {
	mx, my := scn.getMaxXY()
	if r.ll.x < 0 || r.ll.y < 0 || r.ur.x >= mx || r.ur.y >= my {
		return outOfBoundsErr
	}
	if colorUnknown(r.c) {
		return colorUnknownErr
	}

	// Fill every pixel inside the rectangle bounds.
	for x := r.ll.x; x < r.ur.x; x++ {
		for y := r.ll.y; y < r.ur.y; y++ {
			if err := scn.drawPixel(x, y, r.c); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r Rectangle) printShape() string {
	return fmt.Sprintf("Rectangle: (%d,%d) to (%d,%d)", r.ll.x, r.ll.y, r.ur.x, r.ur.y)
}

// Triangle

// Linear interpolation helper.
func interpolate(l0, d0, l1, d1 int) []int {
	a := float64(d1-d0) / float64(l1-l0) // slope
	d := float64(d0)
	n := l1 - l0 + 1
	out := make([]int, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, int(d))
		d += a
	}
	return out
}

func (t Triangle) draw(scn screen) error {
	// Bounds / color checks.
	if outOfBounds(t.pt0, scn) || outOfBounds(t.pt1, scn) || outOfBounds(t.pt2, scn) {
		return outOfBoundsErr
	}
	if colorUnknown(t.c) {
		return colorUnknownErr
	}

	// Sort vertices by ascending y to simplify scan-line fill.
	x0, y0 := t.pt0.x, t.pt0.y
	x1, y1 := t.pt1.x, t.pt1.y
	x2, y2 := t.pt2.x, t.pt2.y
	if y1 < y0 {
		x0, y0, x1, y1 = x1, y1, x0, y0
	}
	if y2 < y0 {
		x0, y0, x2, y2 = x2, y2, x0, y0
	}
	if y2 < y1 {
		x1, y1, x2, y2 = x2, y2, x1, y1
	}

	// Interpolate edges.
	x01 := interpolate(y0, x0, y1, x1)
	x12 := interpolate(y1, x1, y2, x2)
	x02 := interpolate(y0, x0, y2, x2)
	x012 := append(x01[:len(x01)-1], x12...)
	mid := len(x012) / 2

	// Determine left/right edges for each scan line.
	left, right := x012, x02
	if x02[mid] < x012[mid] {
		left, right = x02, x012
	}

	// Fill horizontal spans.
	for yy := y0; yy <= y2; yy++ {
		for xx := left[yy-y0]; xx <= right[yy-y0]; xx++ {
			if err := scn.drawPixel(xx, yy, t.c); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t Triangle) printShape() string {
	return fmt.Sprintf("Triangle: (%d,%d), (%d,%d), (%d,%d)",
		t.pt0.x, t.pt0.y, t.pt1.x, t.pt1.y, t.pt2.x, t.pt2.y)
}

// Circle

// True if tile lies inside or exactly on the circle.
func insideCircle(center, tile Point, r float64) bool {
	dx := float64(center.x - tile.x)
	dy := float64(center.y - tile.y)
	return math.Sqrt(dx*dx+dy*dy) <= r
}

func (circ Circle) draw(scn screen) error {
	// Bounds / color checks.
	if circ.center.x-circ.r < 0 || circ.center.y-circ.r < 0 {
		return outOfBoundsErr
	}
	mx, my := scn.getMaxXY()
	if circ.center.x+circ.r >= mx || circ.center.y+circ.r >= my {
		return outOfBoundsErr
	}
	if colorUnknown(circ.c) {
		return colorUnknownErr
	}

	height := circ.center.y + circ.r
	width := circ.center.x + circ.r
	// Scan entire bounding square; draw if tile is inside circle.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if insideCircle(circ.center, Point{x, y}, float64(circ.r)) {
				scn.drawPixel(x, y, circ.c)
			}
		}
	}
	return nil
}

func (circ Circle) printShape() string {
	return fmt.Sprintf("Circle: centered around (%d,%d) with radius %d",
		circ.center.x, circ.center.y, circ.r)
}

// Display (screen implementation)

func (d *Display) initialize(x, y int) {
	d.maxX, d.maxY = x, y
	d.matrix = make([][]Color, x)
	for i := range d.matrix {
		d.matrix[i] = make([]Color, y)
	}
	d.clearScreen() // set everything to white
}

func (d *Display) getMaxXY() (int, int) { return d.maxX, d.maxY }

func (d *Display) drawPixel(x, y int, c Color) error {
	if x < 0 || x >= d.maxX || y < 0 || y >= d.maxY {
		return outOfBoundsErr
	}
	if colorUnknown(c) {
		return colorUnknownErr
	}
	d.matrix[x][y] = c
	return nil
}

func (d *Display) getPixel(x, y int) (Color, error) {
	if x < 0 || x >= d.maxX || y < 0 || y >= d.maxY {
		return "", outOfBoundsErr
	}
	colr := d.matrix[x][y]
	if colorUnknown(colr) {
		return "", colorUnknownErr
	}
	return colr, nil
}

func (d *Display) clearScreen() {
	for r := range d.matrix {
		for c := range d.matrix[r] {
			d.matrix[r][c] = "white" // reset to white
		}
	}
}

func (d *Display) screenShot(f string) error {
	file, err := os.Create(f + ".ppm")
	if err != nil {
		fmt.Println("**Error creating ppm file:", err)
		return err
	}
	defer file.Close()

	// Write PPM header.
	fmt.Fprintln(file, "P3")
	fmt.Fprintf(file, "%d %d\n", d.maxX, d.maxY)
	fmt.Fprintln(file, "255")

	// Write pixel data row by row.
	for r := 0; r < d.maxX; r++ {
		for c := 0; c < d.maxY; c++ {
			rgb := colorMap[d.matrix[r][c]]
			fmt.Fprintf(file, "%d %d %d ", rgb[0], rgb[1], rgb[2])
		}
		fmt.Fprintln(file)
	}
	return nil
}
