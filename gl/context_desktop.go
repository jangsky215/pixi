package gl

import (
	"fmt"
	"image"
	"image/draw"
	"reflect"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/pkg/errors"
)

type context struct {
}

func Init() error {
	if err := gl.Init(); err != nil {
		return err
	}
	theContext = &Context{}

	return nil
}

func (c *context) callNonBlock(f func()) {
	f()
}

/*
 *	Buffer
 */
type Buffer struct {
	glid   uint32
	gltype uint32
	size   int
}

func newBuffer(gltype uint32, slice interface{}) *Buffer {
	buffer := &Buffer{
		gltype: gltype,
	}
	gl.GenBuffers(1, &buffer.glid)

	if slice != nil {
		buffer.update(gl.STATIC_DRAW, slice)
	}

	return buffer
}

func NewVertexBuffer(slice interface{}) *Buffer {
	return newBuffer(gl.ARRAY_BUFFER, slice)
}

func NewIndexBuffer(slice interface{}) *Buffer {
	return newBuffer(gl.ELEMENT_ARRAY_BUFFER, slice)
}

func (buffer *Buffer) Destroy() {
	gl.DeleteBuffers(1, &buffer.glid)
}

func (buffer *Buffer) update(drawType uint32, slice interface{}) {
	val := reflect.ValueOf(slice)
	if val.Kind() != reflect.Slice {
		panic(errors.New("expected slice"))
	}
	size := val.Len() * int(val.Type().Elem().Size())
	gl.BindBuffer(buffer.gltype, buffer.glid)
	if buffer.size >= size {
		gl.BufferSubData(buffer.gltype, 0, size, gl.Ptr(slice))
	} else {
		buffer.size = size
		gl.BufferData(buffer.gltype, size, gl.Ptr(slice), drawType)
	}
}

func (buffer *Buffer) Upload(slice interface{}) {
	buffer.update(gl.STREAM_DRAW, slice)
}

/*
 *	Texture
 */
type Texture struct {
	glid          uint32
	width, height int
	mipmap        bool
}

func NewTexture() *Texture {
	tex := &Texture{}
	gl.GenTextures(1, &tex.glid)
	return tex
}

func (tex *Texture) Destroy() {
	gl.DeleteTextures(1, &tex.glid)
}

func (tex *Texture) Bind() {
	gl.BindTexture(gl.TEXTURE_2D, tex.glid)
}

func (tex *Texture) ActiveTexture(i int) {
	gl.ActiveTexture(gl.TEXTURE_2D + uint32(i))
	tex.Bind()
}

func (tex *Texture) EnableMipmap() {
	tex.Bind()
	tex.mipmap = true
	gl.GenerateMipmap(gl.TEXTURE_2D)
}

func (tex *Texture) UploadImg(img image.Image) {
	var rgba *image.RGBA
	if t, ok := img.(*image.RGBA); ok {
		rgba = t
	} else {
		rgba = image.NewRGBA(img.Bounds())
		if rgba.Stride != rgba.Rect.Size().X*4 {
			panic(errors.New("unsupported stride"))
		}
		draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)
	}
	tex.Upload(rgba.Pix, rgba.Rect.Size().X, rgba.Rect.Size().Y)
}

func (tex *Texture) Upload(pixels []uint8, width, height int) {
	var ptr unsafe.Pointer
	if pixels != nil {
		ptr = gl.Ptr(pixels)
	}

	tex.Bind()
	if tex.mipmap {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	} else {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	}
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)

	if tex.width != width || tex.height != height {
		tex.width = width
		tex.height = height
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, ptr)
	} else {
		gl.TexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, int32(width), int32(height), gl.RGBA, gl.UNSIGNED_BYTE, ptr)
	}
}

/*
 *	Framebuffer
 */
type Framebuffer struct {
	glid          uint32
	stencil       uint32
	width, height int
	tex           *Texture
}

func NewFramebuffer(width, height int) *Framebuffer {
	fb := &Framebuffer{
		width:  width,
		height: height,
		tex:    NewTexture(),
	}
	gl.CreateFramebuffers(1, &fb.glid)
	fb.Bind()

	fb.tex.Upload(nil, width, height)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, fb.tex.glid, 0)

	return fb
}

func (fb *Framebuffer) Destroy() {
	gl.DeleteFramebuffers(1, &fb.glid)
	if fb.stencil > 0 {
		gl.DeleteFramebuffers(1, &fb.stencil)
	}
	fb.tex.Destroy()
}

func (fb *Framebuffer) EnableStencil() {
	if fb.stencil > 0 {
		return
	}
	gl.CreateFramebuffers(1, &fb.stencil)
	gl.BindRenderbuffer(gl.RENDERBUFFER, fb.stencil)

	gl.FramebufferRenderbuffer(gl.RENDERBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, fb.stencil)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_STENCIL, int32(fb.width), int32(fb.height))
}

func (fb *Framebuffer) Bind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.glid)
}

func (fb *Framebuffer) Unbind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func (fb *Framebuffer) Clear(r, g, b, a float32) {
	fb.Bind()
	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (fb *Framebuffer) Resize(width, height int) {
	fb.width = width
	fb.height = height

	fb.tex.Upload(nil, width, height)

	if fb.stencil > 0 {
		gl.BindRenderbuffer(gl.RENDERBUFFER, fb.stencil)
		gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_STENCIL, int32(fb.width), int32(fb.height))
	}
}

/*
 *	Shader
 */
type Shader struct {
	glid        uint32
	attribs     []Attr
	uniforms    []Attr
	uniformsLoc []int32
	samplers    []int32
}

func compileShader(shaderType uint32, source string) uint32 {
	shader := gl.CreateShader(shaderType)

	cSrc, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, cSrc, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		panic(errors.Errorf("failed to compile %v: %v", source, log))
	}
	return shader
}

func NewShader(vertexSrc, fragmentSrc string, attribs []Attr, uniforms []Attr) *Shader {
	vertShader := compileShader(gl.VERTEX_SHADER, vertexSrc)
	fragShader := compileShader(gl.FRAGMENT_SHADER, fragmentSrc)

	program := gl.CreateProgram()
	gl.AttachShader(program, vertShader)
	gl.AttachShader(program, fragShader)

	gl.DeleteShader(vertShader)
	gl.DeleteShader(fragShader)

	shader := &Shader{
		glid:     program,
		attribs:  attribs,
		uniforms: uniforms,
	}

	for i, attr := range attribs {
		//必须在 LinkProgram 之前
		gl.BindAttribLocation(program, uint32(i), gl.Str(attr.Name+"\x00"))
	}

	gl.LinkProgram(program)
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		panic(fmt.Errorf("failed to link program: %v", log))
	}

	for _, uniform := range uniforms {
		loc := gl.GetUniformLocation(program, gl.Str(uniform.Name+"\x00"))
		shader.uniformsLoc = append(shader.uniformsLoc, loc)
		if uniform.Type == Sampler2D {
			shader.samplers = append(shader.samplers, loc)
		}
	}

	return shader
}

func (shader *Shader) getAttrib() {
	program := shader.glid

	var count int32
	gl.GetProgramiv(program, gl.ACTIVE_ATTRIBUTES, &count)

	var length, maxLength int32
	gl.GetProgramiv(program, gl.ACTIVE_ATTRIBUTE_MAX_LENGTH, &maxLength)

	data := make([]uint8, maxLength)
	var xtype uint32

	for i := int32(0); i < count; i++ {
		gl.GetActiveAttrib(program, uint32(i), maxLength, &length, nil, &xtype, &data[0])
		loc := gl.GetAttribLocation(program, &data[0])

		name := string(data[:length])
		fmt.Println(name, loc == i, xtype == gl.SAMPLER_2D)
	}
}

func (shader *Shader) Bind() {
	gl.UseProgram(shader.glid)
}

func (shader *Shader) Destroy() {
	gl.DeleteShader(shader.glid)
}

/*
 *	VertexArrayObject
 */
type VertexArrayObject struct {
	glid        uint32
	indexBuffer *Buffer
	shader      *Shader

	dirty bool
}

func NewVertexArrayObject() *VertexArrayObject {
	vao := &VertexArrayObject{}
	gl.GenVertexArrays(1, &vao.glid)

	return vao
}

func (vao *VertexArrayObject) Destroy() {
	gl.DeleteVertexArrays(1, &vao.glid)
}

func (vao *VertexArrayObject) SetIndexBuffer(indexBuffer *Buffer) {
	vao.dirty = true
	vao.indexBuffer = indexBuffer
}

func (vao *VertexArrayObject) SetShader(shader *Shader) {
	vao.dirty = true
	vao.shader = shader
}

func (vao *VertexArrayObject) Draw(mode DrawMode, count, start int) {
	gl.DrawElements(uint32(mode), int32(count), gl.UNSIGNED_SHORT, unsafe.Pointer(uintptr(start)))
}
