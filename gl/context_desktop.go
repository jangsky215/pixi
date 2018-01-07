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

	stride      int32
	attrLayouts []attrLayout
}

type attrLayout struct {
	name       string
	num        int32
	xtype      uint32
	normalized bool
	pointer    unsafe.Pointer
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

func NewVertexBuffer(slice interface{}, attrs Attrs) *Buffer {
	buffer := newBuffer(gl.ARRAY_BUFFER, slice)

	buffer.attrLayouts = make([]attrLayout, len(attrs))
	offset := uintptr(0)
	for i, attr := range attrs {
		layout := attrLayout{
			name:       attr.Name,
			num:        int32(attr.Num),
			xtype:      uint32(attr.Type),
			normalized: attr.Type.normalized(),
			pointer:    unsafe.Pointer(offset),
		}
		offset += uintptr(attr.Type.size() * attr.Num)
		buffer.attrLayouts[i] = layout
	}
	buffer.stride = int32(offset)

	return buffer
}

func NewIndexBuffer(slice interface{}) *Buffer {
	return newBuffer(gl.ELEMENT_ARRAY_BUFFER, slice)
}

func (buffer *Buffer) Destroy() {
	gl.DeleteBuffers(1, &buffer.glid)
}

func (buffer *Buffer) bind() {
	gl.BindBuffer(buffer.gltype, buffer.glid)
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
	if tex.mipmap {
		return
	}
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
	glid       uint32
	attributes Attributes
	uniforms   Attributes
	samplers   []int32
}

type Attributes map[string]int32

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

func NewShader(vertexSrc, fragmentSrc string) *Shader {
	vertShader := compileShader(gl.VERTEX_SHADER, vertexSrc)
	fragShader := compileShader(gl.FRAGMENT_SHADER, fragmentSrc)

	program := gl.CreateProgram()
	gl.AttachShader(program, vertShader)
	gl.AttachShader(program, fragShader)

	gl.DeleteShader(vertShader)
	gl.DeleteShader(fragShader)

	shader := &Shader{
		glid:       program,
		attributes: make(Attributes),
		uniforms:   make(Attributes),
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

	shader.getAttributes()
	shader.getUniforms()

	return shader
}

func (shader *Shader) getAttributes() {
	program := shader.glid

	var count int32
	gl.GetProgramiv(program, gl.ACTIVE_ATTRIBUTES, &count)

	var length, maxLength int32
	gl.GetProgramiv(program, gl.ACTIVE_ATTRIBUTE_MAX_LENGTH, &maxLength)

	data := make([]uint8, maxLength)

	for i := int32(0); i < count; i++ {
		gl.GetActiveAttrib(program, uint32(i), maxLength, &length, nil, nil, &data[0])
		loc := gl.GetAttribLocation(program, &data[0])
		name := string(data[:length])
		shader.attributes[name] = loc
	}
}

func (shader *Shader) getUniforms() {
	program := shader.glid

	var count int32
	gl.GetProgramiv(program, gl.ACTIVE_UNIFORMS, &count)

	var length, maxLength int32
	gl.GetProgramiv(program, gl.ACTIVE_UNIFORM_MAX_LENGTH, &maxLength)

	data := make([]uint8, maxLength)
	var xtype uint32

	for i := int32(0); i < count; i++ {
		gl.GetActiveUniform(program, uint32(i), maxLength, &length, nil, &xtype, &data[0])
		loc := gl.GetUniformLocation(program, &data[0])
		name := string(data[:length])
		shader.uniforms[name] = loc
		if xtype == gl.SAMPLER_2D {
			shader.samplers = append(shader.samplers, loc)
		}
	}
}

func (shader *Shader) Bind() {
	gl.UseProgram(shader.glid)
}

func (shader *Shader) Destroy() {
	gl.DeleteShader(shader.glid)
}

// Attrib
func (shader *Shader) Attributes() Attributes {
	return shader.attributes
}

// Uniform
func (shader *Shader) SetSampler2D(index, offset int) {
	//绑定纹理目标
	gl.Uniform1i(shader.samplers[index], int32(offset)) // gl.TEXTURE0 + offset
}

func (shader *Shader) UniformLocation(name string) int32 {
	return shader.uniforms[name]
}

func (shader *Shader) SetUniformName(name string, v ...float32) {
	loc, exist := shader.uniforms[name]
	if !exist {
		panic("name not exist")
	}
	shader.SetUniform(loc, v...)
}

func (shader *Shader) SetUniform(loc int32, v ...float32) {
	switch len(v) {
	case 1: //gl.FLOAT:
		gl.Uniform1f(loc, v[0])
	case 2: //gl.FLOAT_VEC2:
		gl.Uniform2f(loc, v[0], v[1])
	case 3: //gl.FLOAT_VEC3:
		gl.Uniform3f(loc, v[0], v[1], v[2])
	case 4: //gl.FLOAT_VEC4:
		gl.Uniform4f(loc, v[0], v[1], v[2], v[3])
	case 9: //gl.FLOAT_MAT3:
		gl.UniformMatrix3fv(loc, 1, false, &v[0])
	case 16: //gl.FLOAT_MAT4:
		gl.UniformMatrix4fv(loc, 1, false, &v[0])
	default:
		panic("error uniform type")
	}
}

/*
 *	VertexArrayObject
 */
type VertexArrayObject struct {
	glid          uint32
	attributes    Attributes
	indexBuffer   *Buffer
	vertexBuffers []*Buffer

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

func (vao *VertexArrayObject) SetAttributes(attributes Attributes) {
	vao.dirty = true
	vao.attributes = attributes
}

func (vao *VertexArrayObject) AddBuffer(buffer *Buffer) {
	vao.dirty = true
	vao.vertexBuffers = append(vao.vertexBuffers, buffer)
}

func (vao *VertexArrayObject) activate() {
	for _, buffer := range vao.vertexBuffers {
		buffer.bind()
		for _, al := range buffer.attrLayouts {
			loc, exist := vao.attributes[al.name]
			if !exist {
				panic(fmt.Errorf(`attribute name "%s" not exist`, al.name))
			}
			index := uint32(loc)
			gl.EnableVertexAttribArray(index)
			gl.VertexAttribPointer(index, al.num, al.xtype, al.normalized, buffer.stride, al.pointer)
		}
	}

	if vao.indexBuffer != nil {
		vao.indexBuffer.bind()
	}
}

func (vao *VertexArrayObject) Bind() {
	gl.BindVertexArray(vao.glid)
	if vao.dirty {
		vao.dirty = false
		vao.activate()
	}
}

func (vao *VertexArrayObject) Unbind() {
	gl.BindVertexArray(0)
}

func (vao *VertexArrayObject) Clear() {
	vao.dirty = true
	vao.indexBuffer = nil
	vao.vertexBuffers = nil
	vao.attributes = nil
}

func (vao *VertexArrayObject) IndexBuffer() *Buffer {
	return vao.indexBuffer
}

func (vao *VertexArrayObject) VertexBuffers() []*Buffer {
	return vao.vertexBuffers
}

func (vao *VertexArrayObject) Draw(mode DrawMode, start, count int) {
	if vao.indexBuffer != nil {
		gl.DrawElements(uint32(mode), int32(count), gl.UNSIGNED_SHORT, unsafe.Pointer(uintptr(start)))
	} else {
		gl.DrawArrays(uint32(mode), int32(start), int32(count))
	}
}

/*
 *	Util
 */
func Clear(r, g, b, a float32) {
	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func Viewport(x, y, width, height int) {
	gl.Viewport(int32(x), int32(y), int32(width), int32(height))
}
