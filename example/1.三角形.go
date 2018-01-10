package main

import (
	"runtime"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	"github.com/jangsky215/pixi/gl"
)

func main() {
	runtime.LockOSThread()

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

	if err := gl.Init(); err != nil {
		panic(err)
	}

	s := gl.NewShader(vertexShader, fragmentShader)
	s.Bind()

	vao := gl.NewVertexArrayObject()

	vertexBuffer := gl.NewVertexBuffer(points, gl.Attrs{
		{"vp", 3, gl.Float},
	})
	vao.AddBuffer(vertexBuffer)

	indexBuffer := gl.NewIndexBuffer(index)
	vao.SetIndexBuffer(indexBuffer)

	vao.SetAttributes(s.Attributes())
	vao.Bind()

	for !window.ShouldClose() {
		gl.Clear(1, 1, 1, 1)
		vao.Draw(gl.DrawTriangle, 0, 3)

		window.SwapBuffers()
		glfw.PollEvents()
	}
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
void main() {
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
