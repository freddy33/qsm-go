package space_nk

import (
	"github.com/xlab/closer"
	"time"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"log"
	"github.com/freddy33/qsm-go/space_gl"
)

const (
	winWidth         = 515
	winHeight        = 544
	maxVertexBuffer  = 100 * 1024 * 1024
	maxElementBuffer = 100 * 1024 * 1024
)

func DisplayNkScatter() {
	win, err := space_gl.InitWindow(winWidth, winHeight, "MDB Plotter Otter")
	if err != nil {
		closer.Fatalln(err)
	}
	win.SetSizeLimits(winWidth, winHeight, winWidth, winHeight)
	err = space_gl.InitGl(win)
	if err != nil {
		closer.Fatalln("opengl: initialisation failed:", err)
	}

	ctx := nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)

	// Fonts
	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	sansFont := nk.NkFontAtlasAddDefault(atlas, 16, nil)
	nk.NkFontStashEnd()
	if sansFont != nil {
		nk.NkStyleSetFont(ctx, sansFont.Handle())
	}
	exitC := make(chan struct{}, 1)
	doneC := make(chan struct{}, 1)
	closer.Bind(func() {
		close(exitC)
		<-doneC
	})
	fpsTicker := time.NewTicker(time.Second / 30)
	for {
		select {
		case <-exitC:
			nk.NkPlatformShutdown()
			glfw.Terminate()
			fpsTicker.Stop()
			close(doneC)
			return
		case <-fpsTicker.C:
			if win.ShouldClose() {
				close(exitC)
				continue
			}
			glfw.PollEvents()
			gfxMain(win, ctx)

		}
	}
}

func plotWidget(ctx *nk.Context, canvas *nk.CommandBuffer) {
	state := nk.NkWidget(nk.NewRect(), ctx)
	switch state {
	case nk.WidgetInvalid:
		return
	case nk.WidgetRom:
		// temporary state
	case nk.WidgetValid:
		// update by user input
	}
	input := ctx.Input()
	for fy := 50; fy <= 400; fy += 10 {
		for fx := 50; fx <= 400; fx += 10 {
			c1 := nk.NkRect(float32(fx), float32(fy), 5.0, 5.0)
			nk.NkFillCircle(canvas, c1, nk.NkRgb(171, 239, 29))
			if nk.NkInputHasMouseClickDownInRect(input, nk.ButtonLeft, c1, 1) > 0 {
				log.Println("Receive a click on coordinate: ", fx, fy)
			}
		}
	}
}

func gfxMain(win *glfw.Window, ctx *nk.Context) {
	nk.NkPlatformNewFrame()
	plotAreaWidth := float32(winWidth)
	plotAreaHeight := float32(winHeight)

	// Layout
	bounds := nk.NkRect(0, 0, plotAreaWidth, plotAreaHeight)
	if nk.NkBegin(ctx, " ", bounds, nk.WindowTitle|nk.WindowNoScrollbar) > 0 {
		canvas := nk.NkWindowGetCanvas(ctx)
		totalSpace := nk.NkWindowGetContentRegion(ctx)
		nk.NkLayoutSpaceBegin(ctx, nk.LayoutStatic, totalSpace.H(), 10000)
		//Grid
		gridColor := nk.NkRgb(80, 80, 80)
		gridSize := float32(128.0)
		winSize := nk.NkLayoutSpaceBounds(ctx)
		for x := float32(0.0); x < winSize.W(); x += gridSize {
			nk.NkStrokeLine(canvas, x+winSize.X(), winSize.Y(), x+winSize.X(), winSize.Y()+winSize.H(), 1.0, gridColor)
		}
		for y := float32(0.0); y < winSize.H(); y += gridSize {
			nk.NkStrokeLine(canvas, winSize.X(), y+winSize.Y(), winSize.X()+winSize.W(), y+winSize.Y(), 1.0, gridColor)
		}
		nk.NkLayoutSpaceEnd(ctx)
		plotWidget(ctx, canvas)
	}
	nk.NkEnd(ctx)

	// Render
	bg := make([]float32, 4)
	nk.NkColorFv(bg, nk.NkRgba(50, 50, 50, 255))
	width, height := win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
}