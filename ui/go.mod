module github.com/freddy33/qsm-go/ui

require (
	github.com/freddy33/qsm-go/model v0.0.0-20200626140801-6f9a15bc7381
	github.com/freddy33/qsm-go/m3util v0.0.0-latest
	github.com/go-gl/gl v0.0.0-20190320180904-bf2b1f2f34d7
	github.com/go-gl/glfw v0.0.0-20191125211704-12ad95a8df72
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20200222043503-6f7a984d4dc4
	github.com/go-gl/mathgl v0.0.0-20190713194549-592312d8590a
	github.com/stretchr/testify v1.3.0
	golang.org/x/image v0.0.0-20200119044424-58c23975cae1 // indirect
)

replace github.com/freddy33/qsm-go/m3util => ../m3util
replace github.com/freddy33/qsm-go/model => ../model

go 1.13
