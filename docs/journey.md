# Journey to develop a go version of QSM

# Using gonum, gonum plot and glot
The go and science community are gathering around [gonum](https://gonum.org/) and it has it's own plotting repository [gonum/plot](https://github.com/gonum/plot) with full doc [here](https://godoc.org/gonum.org/v1/plot).

By googling a little more I found [glot](https://medium.com/@Arafat./introducing-glot-the-plotting-library-for-golang-3133399948a1) with [source here](https://github.com/Arafatk/glot).

I played around with it for a bit and reached this:
![]()

It was good for the target (3D plotting of data) but not suited at all for QSM. Missing interactive behavior, zoom and rotate, and most important all the points and nodes are not a single plot line.
So I looked for something else.

# The Nuklear way
Looking for Desktop UI 3D with go I found [nuklear](https://github.com/vurtun/nuklear) with [go binding here](https://github.com/golang-ui/nuklear)


First 
