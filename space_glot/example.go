package space_glot

import "github.com/Arafatk/glot"

func DisplayExample3D() {
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
