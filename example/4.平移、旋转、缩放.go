package main

import (
	"runtime"

	glfw "github.com/go-gl/glfw/v3.1/glfw"
	pixi "github.com/jangsky215/pixi/gl"
	"github.com/jangsky215/pixi/math"
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

	width, height := 800, 600
	window, err := glfw.CreateWindow(width, height, "Tutorial #1", nil, nil)

	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	if err := pixi.Init(); err != nil {
		panic(err)
	}

	s := pixi.NewShader(vertexShader, fragmentShader)
	s.Bind()

	vao := pixi.NewVertexArrayObject()

	vertexBuffer := pixi.NewVertexBuffer(nil, pixi.Attrs{
		{"vp", 3, pixi.Float},
	})
	vao.AddBuffer(vertexBuffer)

	indexBuffer := pixi.NewIndexBuffer(index)
	vao.SetIndexBuffer(indexBuffer)

	vao.SetAttributes(s.Attributes())
	vao.Bind()

	aspect := float32(width) / float32(height) // = glheight / glwidth
	angle := float32(0)
	for !window.ShouldClose() {
		pixi.Clear(1, 1, 1, 1)

		angle += 0.5
		m := &math.Matrix{}
		m.Identity()
		m.Scale(0.5, 0.5)
		m.Translate(0.5, 0.5)
		m.Rotate(angle * math.RadianFactor)

		vertex := make([]float32, len(points))
		copy(vertex, points)
		for i := 0; i < len(vertex); i += 3 {
			vertex[i], vertex[i+1] = m.Apply(vertex[i], vertex[i+1])
			vertex[i+1] *= aspect
		}
		vertexBuffer.Upload(vertex)

		vao.Draw(pixi.DrawTriangle, 0, 3)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

var points = []float32{
	0.0, 0.5, 0.0,
	0.5, 0.0, 0.0,
	-0.5, 0.0, 0.0,
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
