package main

import (
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	glfw "github.com/go-gl/glfw/v3.1/glfw"
	pixi "github.com/jangsky215/pixi/gl"
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()
}

var points = []float32{
	0.0, 0.5, 0.0,
	0.5, -0.5, 0.0,
	-0.5, -0.5, 0.0,
}

var index = []int16{
	0, 1, 2,
}

var vertexShader = `
#version 410

in vec3 vp;
in float vp1;
void main() {
if(vp1==0){

}
gl_Position = vec4(vp, 1.0);
}
` + "\x00"

var fragmentShader = `
#version 410

out vec4 frag_colour;
void main() {
frag_colour = vec4(0.5, 1.0, 0.5, 1.0);
}
` + "\x00"

func main() {
	runtime.LockOSThread()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)    // Necessary for OS X
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile) // Necessary for OS X
	//glfw.WindowHint(glfw.OpenGLDebugContext, glfw.True)

	glfw.WindowHint(glfw.Resizable, glfw.True)

	window, err := glfw.CreateWindow(600, 480, "Tutorial #1", nil, nil)

	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	s := pixi.NewShader(vertexShader, fragmentShader)
	s.Bind()

	vao := pixi.NewVertexArrayObject()

	vertexBuffer := pixi.NewVertexBuffer(points, pixi.Attrs{
		{"vp", 3, pixi.Float},
	})
	vao.AddBuffer(vertexBuffer)

	indexBuffer := pixi.NewIndexBuffer(index)
	vao.SetIndexBuffer(indexBuffer)

	vao.SetAttributes(s.GetAttributes())
	vao.Bind()

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		vao.Draw(pixi.DrawTriangle, 0, 3)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
