package internal

import (
	"errors"
	"fmt"
	"image/draw"
	"runtime"

	"image"
	"reflect"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type context struct {
}

func Init() error {
	if err := gl.Init(); err != nil {
		return err
	}
	ctx := newContext()

	theContext = ctx

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

	stride int32
}

func newBuffer(gltype uint32, slice interface{}, stride int32) *Buffer {
	buffer := &Buffer{
		gltype: gltype,
		stride: stride,
	}
	gl.GenBuffers(1, &buffer.glid)

	if slice != nil {
		buffer.update(gl.STATIC_DRAW, slice)
	}

	runtime.SetFinalizer(buffer, (*Buffer).delete)

	return buffer
}

func NewVertexBuffer(slice interface{}, stride int32) *Buffer {
	return newBuffer(gl.ARRAY_BUFFER, slice, stride)
}

func NewIndexBuffer(slice []uint16) *Buffer {
	return newBuffer(gl.ELEMENT_ARRAY_BUFFER, slice, 0)
}

func (buffer *Buffer) delete() {
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
	gl.BindVertexArray(0)
	buffer.bind()
	gl.BufferData(buffer.gltype, size, gl.Ptr(slice), drawType)
}

func (buffer *Buffer) Upload(slice interface{}) {
	buffer.update(gl.STREAM_DRAW, slice)
}

/*
 *	Texture
 */
type Texture struct {
	glid   uint32
	mipmap bool
}

func NewTexture() *Texture {
	tex := &Texture{}
	gl.GenTextures(1, &tex.glid)

	runtime.SetFinalizer(tex, (*Texture).delete)

	return tex
}

func (tex *Texture) delete() {
	gl.DeleteTextures(1, &tex.glid)
}

func (tex *Texture) bind() {
	gl.ActiveTexture(gl.TEXTURE7)
	gl.BindTexture(gl.TEXTURE_2D, tex.glid)
}

func (tex *Texture) activeTexture(i int) {
	gl.ActiveTexture(gl.TEXTURE0 + uint32(i))
	gl.BindTexture(gl.TEXTURE_2D, tex.glid)
}

func (tex *Texture) EnableMipmap() {
	if tex.mipmap {
		return
	}
	tex.bind()
	tex.mipmap = true
	gl.GenerateMipmap(gl.TEXTURE_2D)
}

func (tex *Texture) UploadImage(img image.Image) {
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

	tex.bind()
	if tex.mipmap {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	} else {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	}
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, ptr)
}

func (tex *Texture) SubUpload(pixels []uint8, x, y, width, height int) {
	tex.bind()
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, int32(x), int32(y), int32(width), int32(height),
		gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))
}

/*
 *	Target
 */
type Target struct {
	glid          uint32
	stencil       uint32
	width, height int
	tex           *Texture
}

func NewTarget(width, height int) *Target {
	target := &Target{
		width:  width,
		height: height,
		tex:    NewTexture(),
	}
	gl.GenFramebuffers(1, &target.glid)
	target.bind()
	theContext.dirtyFlag |= dirtyTarget

	target.tex.Upload(nil, width, height)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, target.tex.glid, 0)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("init Framebuffer error")
	}

	runtime.SetFinalizer(target, (*Target).delete)

	return target
}

func (target *Target) delete() {
	gl.DeleteFramebuffers(1, &target.glid)
	if target.stencil > 0 {
		gl.DeleteFramebuffers(1, &target.stencil)
	}
}

func (target *Target) Texture() *Texture {
	return target.tex
}

func (target *Target) EnableStencil() {
	if target.stencil > 0 {
		return
	}
	gl.CreateFramebuffers(1, &target.stencil)
	gl.BindRenderbuffer(gl.RENDERBUFFER, target.stencil)

	gl.FramebufferRenderbuffer(gl.RENDERBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, target.stencil)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_STENCIL, int32(target.width), int32(target.height))
}

func (target *Target) bind() {
	if target != nil {
		gl.BindFramebuffer(gl.FRAMEBUFFER, target.glid)
	} else {
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	}
}

func (target *Target) Clear(r, g, b, a float32) {
	target.bind()
	Clear(r, g, b, a)
}

func (target *Target) Resize(width, height int) {
	target.width = width
	target.height = height

	target.tex.Upload(nil, width, height)

	if target.stencil > 0 {
		gl.BindRenderbuffer(gl.RENDERBUFFER, target.stencil)
		gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_STENCIL, int32(target.width), int32(target.height))
	}
}

/*
 *	Shader
 */
type Shader struct {
	glid            uint32
	glvao           uint32
	attributes      map[string]int32
	uniforms        map[string]int32
	textureUniforms []int32

	attribLayout []layout
	bufferDirty  bool
	vertexBuffer *Buffer
	indexBuffer  *Buffer
}

type layout struct {
	loc        uint32
	num        int32
	xtype      uint32
	normalized bool
	pointer    unsafe.Pointer
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

		panic(fmt.Errorf("failed to compile %v: %v", source, log))
	}
	return shader
}

func NewShader(vertexSrc, fragmentSrc string, attrs Attrs) *Shader {
	vertShader := compileShader(gl.VERTEX_SHADER, vertexSrc)
	fragShader := compileShader(gl.FRAGMENT_SHADER, fragmentSrc)

	program := gl.CreateProgram()
	shader := &Shader{
		glid:         program,
		attributes:   make(map[string]int32),
		uniforms:     make(map[string]int32),
		attribLayout: make([]layout, len(attrs)),
	}
	gl.GenVertexArrays(1, &shader.glvao)

	gl.AttachShader(program, vertShader)
	gl.DeleteShader(vertShader)
	gl.AttachShader(program, fragShader)
	gl.DeleteShader(fragShader)

	offset := uintptr(0)
	for i, attr := range attrs {
		gl.BindAttribLocation(program, uint32(i), gl.Str(attr.Name+"\x00"))
		layout := layout{
			loc:        uint32(i),
			num:        int32(attr.Num),
			xtype:      uint32(attr.Type),
			normalized: attr.Type.normalized(),
			pointer:    unsafe.Pointer(offset),
		}
		offset += uintptr(attr.Type.size() * attr.Num)
		shader.attribLayout[i] = layout
	}

	//BindAttribLocation 必须放在 LinkProgram 之前
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

	runtime.SetFinalizer(shader, (*Shader).delete)

	return shader
}

func (shader *Shader) delete() {
	gl.DeleteShader(shader.glid)
	gl.DeleteVertexArrays(1, &shader.glvao)
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
			shader.textureUniforms = append(shader.textureUniforms, loc)
		}
	}
	if len(shader.textureUniforms) == 1 {
		shader.textureUniforms = nil
	}
}

func (shader *Shader) SetVertexBuffer(vertexBuffer *Buffer) {
	if shader.vertexBuffer != vertexBuffer {
		shader.bufferDirty = true
		shader.vertexBuffer = vertexBuffer
	}
}

func (shader *Shader) VertexBuffer() *Buffer {
	return shader.vertexBuffer
}

func (shader *Shader) SetIndexBuffer(indexBuffer *Buffer) {
	if shader.indexBuffer != indexBuffer {
		shader.bufferDirty = true
		shader.indexBuffer = indexBuffer
	}
}

func (shader *Shader) IndexBuffer() *Buffer {
	return shader.indexBuffer
}

func (shader *Shader) bind() {
	gl.UseProgram(shader.glid)
	shader.applyTextureUniform()
	shader.applyVertex()
}

func (shader *Shader) applyVertex() {
	gl.BindVertexArray(shader.glvao)
	if shader.bufferDirty {
		shader.bufferDirty = false
		shader.vertexBuffer.bind()
		for _, al := range shader.attribLayout {
			gl.EnableVertexAttribArray(al.loc)
			gl.VertexAttribPointer(al.loc, al.num, al.xtype, al.normalized, shader.vertexBuffer.stride, al.pointer)
		}
		shader.indexBuffer.bind()
	}
}

// Uniform
func (shader *Shader) applyTextureUniform() {
	//绑定纹理目标
	for i, loc := range shader.textureUniforms {
		gl.Uniform1i(int32(loc), int32(i))
	}
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
 *	Util
 */
func Clear(r, g, b, a float32) {
	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func Viewport(x, y, width, height int) {
	gl.Viewport(int32(x), int32(y), int32(width), int32(height))
}

func enable(cap CapType) {
	gl.Enable(uint32(cap))
}

func disable(cap CapType) {
	gl.Disable(uint32(cap))
}

func blendFunc(src, dst BlendFormat) {
	gl.BlendFunc(uint32(src), uint32(dst))
}

func depthFunc(xfunc DepthFormat) {
	gl.DepthFunc(uint32(xfunc))
}

func depthMask(flag bool) {
	gl.DepthMask(flag)
}

func glDraw(mode DrawMode, start, count int) {
	gl.DrawElements(uint32(mode), int32(count), gl.UNSIGNED_SHORT, unsafe.Pointer(uintptr(start)))
}
