package main

import (
	"github.com/Arafatk/glot"
	"github.com/freddy33/qsm-go/m3"
	"fmt"
)

var space m3.Space

func main() {
	space = m3.Space{}
	var points [16]*m3.Point
	i := 0
	points[i] = &m3.Point{0, 0, 0}
	i++

	points[i] = &m3.Point{1, 1, 0}
	i++
	points[i] = &m3.Point{0, -1, 1}
	i++
	points[i] = &m3.Point{-1, 0, -1}
	i++

	points[i] = &m3.Point{2, 0, -1}
	i++
	points[i] = &m3.Point{0, 2, 1}
	i++

	points[i] = &m3.Point{1, -2, 0}
	i++
	points[i] = &m3.Point{-1, 0, 2}
	i++

	points[i] = &m3.Point{-2, 1, 0}
	i++
	points[i] = &m3.Point{0, -1, -2}
	i++

	points[i] = &m3.Point{3, 0, 0}
	i++
	points[i] = &m3.Point{0, 3, 0}
	i++
	points[i] = &m3.Point{0, -3, 0}
	i++
	points[i] = &m3.Point{0, 0, 3}
	i++
	points[i] = &m3.Point{-3, 0, 0}
	i++
	points[i] = &m3.Point{0, 0, -3}
	i++

	var nodes [len(points)]*m3.Node
	for pos, point := range points {
		nodes[pos] = &m3.Node{P: point}
	}
	nodes[0].C[0] = nodes[1]
	nodes[0].C[1] = nodes[2]
	nodes[0].C[2] = nodes[3]

	for l := 0; l < 3; l++ {
		nodes[l+1].C[0] = nodes[0]
		nodes[l+1].C[1] = nodes[2*l+4]
		nodes[l+1].C[2] = nodes[2*l+5]
		nodes[2*l+4].C[0] = nodes[l+1]
		nodes[2*l+5].C[0] = nodes[l+1]
	}
	for l := 4; l < 10; l++ {
		nodes[l].C[1] = nodes[l+6]
		nodes[l+6].C[0] = nodes[l]
	}
	drawNodes := make([]*m3.Node, 7, 7)
	drawNodes[0] = nodes[0]
	for l := 0; l < 6; l++ {
		drawNodes[l+1] = nodes[l+4]
	}
	draw(points[:], drawNodes)
}

func draw(points []*m3.Point, nodes []*m3.Node) {
	plot, err := glot.NewPlot(3, false, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	pointGroupName := "Nodes"
	style := "points"
	var drawPoints [][]int64
	drawPoints = make([][]int64, 3, 3)
	nbPoints := len(points)
	for d := range drawPoints {
		drawPoints[d] = make([]int64, nbPoints, nbPoints)
		for pos, point := range points {
			switch d {
			case 0:
				drawPoints[d][pos] = point.X
			case 1:
				drawPoints[d][pos] = point.Y
			case 2:
				drawPoints[d][pos] = point.Z
			}
		}
	}
	err = plot.AddPointGroup(pointGroupName, style, drawPoints)
	if err != nil {
		fmt.Println(err)
		return
	}
	style = "lines"
	for ni, n := range nodes {
		if n.C[1] == nil {
			continue
		}
		drawLines := make([][]int64, 3, 3)
		for d := 0; d < 3; d++ {
			if n.C[2] != nil {
				drawLines[d] = make([]int64, 5, 5)
			} else {
				drawLines[d] = make([]int64, 3, 3)
			}
			switch d {
			case 0:
				drawLines[d][0] = n.C[0].P.X
				drawLines[d][1] = n.P.X
				drawLines[d][2] = n.C[1].P.X
				if n.C[2] != nil {
					drawLines[d][3] = n.P.X
					drawLines[d][4] = n.C[2].P.X
				}
			case 1:
				drawLines[d][0] = n.C[0].P.Y
				drawLines[d][1] = n.P.Y
				drawLines[d][2] = n.C[1].P.Y
				if n.C[2] != nil {
					drawLines[d][3] = n.P.Y
					drawLines[d][4] = n.C[2].P.Y
				}
			case 2:
				drawLines[d][0] = n.C[0].P.Z
				drawLines[d][1] = n.P.Z
				drawLines[d][2] = n.C[1].P.Z
				if n.C[2] != nil {
					drawLines[d][3] = n.P.Z
					drawLines[d][4] = n.C[2].P.Z
				}
			}
		}
		err = plot.AddPointGroup(fmt.Sprint("N", ni), style, drawLines)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	for axe := 0 ; axe < 3; axe++ {
		drawCartLines := make([][]int64, 3, 3)
		for d := 0; d < 3; d++ {
			drawCartLines[d] = make([]int64, 2, 2)
			if d == axe {
				drawCartLines[d][0] = -3
				drawCartLines[d][1] = 3
			} else {
				drawCartLines[d][0] = 0
				drawCartLines[d][1] = 0
			}
		}
		err = plot.AddPointGroup(fmt.Sprint("Axe", axe), style, drawCartLines)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// A plot type used to make points/ curves and customize and save them as an image.
	plot.SetTitle("Space Plot")
	// Optional: Setting the title of the plot
	plot.SetXLabel("X")
	plot.SetYLabel("Y")
	plot.SetZLabel("Z")
	// Optional: Setting label for X and Y axis
	plot.SetXrange(-10, 10)
	plot.SetYrange(-10, 10)
	plot.SetYrange(-10, 10)
	// Optional: Setting axis ranges
	plot.SavePlot("3.png")
}

func example3d() {
	dimensions := 3
	// The dimensions supported by the plot
	persist := false
	debug := false
	plot, _ := glot.NewPlot(dimensions, persist, debug)
	pointGroupName := "Simple Circles"
	style := "points"
	points := [][]float64{{7, 3, 13, 5.6, 11.1}, {12, 13, 11, 1, 7}, {12, 13, 11, 1, 7}}
	// Adding a point group
	plot.AddPointGroup(pointGroupName, style, points)
	pointGroupName = "Simple Lines"
	style = "lines"
	points = [][]float64{{7, 3, 3, 5.6, 5.6, 7, 7, 9, 13, 13, 9, 9}, {10, 10, 4, 4, 5.4, 5.4, 4, 4, 4, 10, 10, 4}, {10, 10, 4, 4, 5.4, 5.4, 4, 4, 4, 10, 10, 4}}
	plot.AddPointGroup(pointGroupName, style, points)
	// A plot type used to make points/ curves and customize and save them as an image.
	plot.SetTitle("Example Plot")
	// Optional: Setting the title of the plot
	plot.SetXLabel("X-Axis")
	plot.SetYLabel("Y-Axis")
	// Optional: Setting label for X and Y axis
	plot.SetXrange(-2, 18)
	plot.SetYrange(-2, 18)
	// Optional: Setting axis ranges
	plot.SavePlot("2.png")
}
