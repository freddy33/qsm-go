# Journey to develop a go version of QSM

# Using gonum, gonum plot and glot
The go and science community are gathering around [gonum](https://gonum.org/) and it has it's own plotting repository [gonum/plot](https://github.com/gonum/plot) with full doc [here](https://godoc.org/gonum.org/v1/plot).

By googling a little more I found [glot](https://medium.com/@Arafat./introducing-glot-the-plotting-library-for-golang-3133399948a1) with [source here](https://github.com/Arafatk/glot).

I played around with it for a bit and reached this:
![](https://github.com/freddy33/qsm-go/raw/master/docs/SpaceDots.png)

It was good for the target (3D plotting of data) but not suited at all for QSM. Missing interactive behavior, zoom and rotate, and most important all the points and nodes are not a single plot line.
So I looked for something else.

# The Nuklear way
Looking for Desktop UI 3D with go I found [nuklear](https://github.com/vurtun/nuklear) with [go binding here](https://github.com/golang-ui/nuklear)
It looks great for UI controls and will use it probably later.
The example in my space_nk package taken from [here](https://gist.githubusercontent.com/sindbach/a21d93c5f11a24665d9d07c05340bad3/raw/0cbef62dd1fd33ffc194d9e844d60cc3fd78af0f/test_scatter.go) showed how to use Go OpenGL binding.
Since for now, my issue is mainly 3D rendering of the node and event graph, I went into full OpenGL implementation.
Funny enough after migrating Stellarium to Java using Java OpenGL I thought this will be easy...

# The pure OpenGL way
I have to say Go and OpenGL are a great merge and the [golang OpenGL](https://github.com/go-gl) is active and complete. My issue is that OpenGL 4.1 and GLFW are far from I use to know :D
So I migrated a nice cube example to Go from [this](https://stackoverflow.com/questions/24040982/c-opengl-glfw-drawing-a-simple-cube) nice answer from [genpfault](https://stackoverflow.com/users/44729/genpfault) about good OpenGL practice.
Of course it failed with panic from unexpected signal. And I got stuck for a while.
I finally found [some nice OpenGL go samples](https://github.com/alexozer/opengl-samples-golang) from [alexozer](https://github.com/alexozer). It was using shader and texture which way more than I need. But the example worked and gave me confidence it's the correct direction.
Then I found [KyleBanks](https://github.com/KyleBanks) [Conway's game of life implementation in Go OpenGL](https://github.com/KyleBanks/conways-gol) which also used texture and shader.
So, I found my mistake: This business of VBOs and VAOs is totally foreign to me. Wondered if I completely missed it when I used to do OpenGL in the old time... Well no, it's a new thing. Finally found a small great explanation for it from [Dark Photon](https://www.opengl.org/discussion_boards/showthread.php/199807-A-little-confused-on-the-purpose-intent-of-OpenGL-VAOs-VBOs).

# Defining Vertex
So, that's where I am. I need to thing Vertex, and how I want to draw nodes, events and connections in OpenGL so I can defined VAO and call draw functions.
Let's start...


# Projects and infos to learn from
https://github.com/fogleman/ln
https://medium.com/@drgomesp/opengl-and-golang-getting-started-abcd3d96f3db
https://blog.mapbox.com/drawing-antialiased-lines-with-opengl-8766f34192dc

