package main

import (
	"runtime"

	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"os"
	"time"

	"github.com/go-gl/glfw/v3.1/glfw"
	gl "github.com/jangsky215/pixi/internal"
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

	//混合函数 绘制透明纹理
	gl.SetBlend(gl.BlendSrcAlpha, gl.BlendOneMinusSrcAlpha)

	attrs := gl.Attrs{
		{"position", 3, gl.Float},
		{"color", 3, gl.Float},
		{"texCoord", 2, gl.Float},
	}

	vertexBuffer := gl.NewVertexBuffer(vertices, 8*4)
	indexBuffer := gl.NewIndexBuffer(indices)

	s := gl.NewShader(vertShader, fragShader, attrs)
	s.SetVertexBuffer(vertexBuffer)
	s.SetIndexBuffer(indexBuffer)

	normalS := gl.NewShader(normalVertShader, normalFragShader, attrs)
	normalS.SetVertexBuffer(vertexBuffer)
	normalS.SetIndexBuffer(indexBuffer)

	img := loadImg("./.resource/cat.png")
	tex := gl.NewTexture()
	tex.UploadImage(img)
	gl.SetTexture(tex, 0)

	last := time.Now().Unix()

	for !window.ShouldClose() {
		gl.Clear(1, 1, 1, 1)

		if ((time.Now().Unix()-last)/5)%2 == 0 {
			gl.SetShader(s)
		} else {
			gl.SetShader(normalS)
		}
		gl.Draw(0, 6)

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
	{0.5, 0.5, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0},   // 右上
	{0.5, -0.5, 0.0, 0.0, 1.0, 0.0, 1.0, 1.0},  // 右下
	{-0.5, -0.5, 0.0, 0.0, 0.0, 1.0, 0.0, 1.0}, // 左下
	{-0.5, 0.5, 0.0, 1.0, 1.0, 0.0, 0.0, 0.0},  // 左上
}

var indices = []uint16{
	0, 1, 3, // 第一个三角形
	1, 2, 3, // 第二个三角形
}
var normalVertShader = `
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
    TexCoord = vec2(texCoord.x, 1.0 - texCoord.y); //OpenGl 纹理坐标与图片坐标y轴时相反的所以这里取反
}
`

var normalFragShader = `
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
	if(color.a > 0.5){
		color = mix(color, vec4(ourColor, 1.0), 0.5); //混合颜色
	} 
	//else
	//	color.a = 0;
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
