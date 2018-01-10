package main

import (
	"runtime"

	"fmt"
	glfw "github.com/go-gl/glfw/v3.1/glfw"
	pixi "github.com/jangsky215/pixi/gl"
	"github.com/jangsky215/pixi/math"
	"image"
	"image/draw"
	_ "image/png"
	"os"
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

	//混合函数 绘制透明纹理
	pixi.Enable(pixi.Blend)
	pixi.BlendFunc(pixi.SrcAlpha, pixi.OneMinusSrcAlpha)

	s := pixi.NewShader(vertShader, fragShader)
	s.Bind()

	vao := pixi.NewVertexArrayObject()

	vertexBuffer := pixi.NewVertexBuffer(nil, pixi.Attrs{
		{"position", 3, pixi.Float},
		{"color", 3, pixi.Float},
		{"texCoord", 2, pixi.Float},
	})
	vao.AddBuffer(vertexBuffer)

	indexBuffer := pixi.NewIndexBuffer(indices)
	vao.SetIndexBuffer(indexBuffer)

	vao.SetAttributes(s.Attributes())
	vao.Bind()

	img := loadImg("./.resource/cat.png")
	tex := pixi.NewTexture()
	tex.UploadImg(img)

	s.SetSampler2D(0, 0)

	fb := pixi.NewFramebuffer(width/2, height)
	fb.Unbind()

	aspect := float32(width) / float32(height) // = glheight / glwidth
	angle := float32(0)
	for !window.ShouldClose() {
		pixi.Clear(1, 1, 1, 1)

		tex.Bind()

		fb.Bind()
		fb.Clear(1, 0, 0, 1)
		vertexBuffer.Upload(vertices)
		vao.Draw(pixi.DrawTriangle, 0, 6)
		fb.Unbind()

		fb.BindTexture()

		angle += 0.5
		m := &math.Matrix{}
		m.Identity()
		m.Scale(0.2, 0.2)
		m.Translate(0, 0.5)
		m.Rotate(angle * math.RadianFactor)
		//m.Skew(10*math.RadianFactor, 0*math.RadianFactor)

		vertex := make([]Vertex, len(vertices))
		copy(vertex, vertices)
		for i := 0; i < len(vertex); i++ {
			vertex[i].X, vertex[i].Y = m.Apply(vertex[i].X, vertex[i].Y)
			vertex[i].Y *= aspect
		}
		vertexBuffer.Upload(vertex)

		vao.Draw(pixi.DrawTriangle, 0, 6)

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

type Vertex struct {
	X, Y, Z float32
	R, G, B float32
	U, V    float32
}

var vertices = []Vertex{
	{0.5, 0.5, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0},
	{0.5, -0.5, 0.0, 0.0, 1.0, 0.0, 1.0, 1.0},
	{-0.5, -0.5, 0.0, 0.0, 0.0, 1.0, 0.0, 1.0},
	{-0.5, 0.5, 0.0, 1.0, 1.0, 0.0, 0.0, 0.0},
}

var indices = []uint16{
	0, 1, 3, // 第一个三角形
	1, 2, 3, // 第二个三角形
}

var vertShader = `
#version 410
in vec3 position;
in vec3 color;
in vec2 texCoord;

out vec3 ourColor;
out vec2 TexCoord;

void main()
{
    gl_Position = vec4(position, 1.0f);
    ourColor = color;
    TexCoord = texCoord;
}
`

var fragShader = `
#version 410
in vec3 ourColor;
in vec2 TexCoord;

out vec4 color;

uniform sampler2D ourTexture;

void main()
{
    color = texture(ourTexture, TexCoord); //显示纹理
    //color = vec4(ourColor, 1.0); //显示颜色
	//color = mix(texture(ourTexture, TexCoord), vec4(ourColor, 1.0), 0.5); //混合颜色
}
`

func loadImg(file string) *image.RGBA {
	imgFile, err := os.Open(file)
	if err != nil {
		panic(fmt.Errorf("texture %q not found on disk: %v", file, err))
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		panic(err)
	}

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		panic(fmt.Errorf("unsupported stride"))
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	return rgba
}

