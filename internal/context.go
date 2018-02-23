package internal

type DirtyFlag uint32

const (
	dirtyTexture DirtyFlag = 1 << iota
	dirtyBlend
	dirtyDepth
	dirtyTarget
	dirtyScissor
)

type Context struct {
	context
	dirtyFlag          DirtyFlag
	attrs              Attrs
	blendSrc, blendDst BlendFormat
	depth              DepthFormat
	depthmask          bool
	scissor            bool
	shader             *Shader
	texture            [8]*Texture
	target             *Target
}

func newContext() *Context {
	return &Context{}
}

var theContext *Context

func GetContext() *Context {
	return theContext
}

func SetAttrs(attrs Attrs) {
	theContext.attrs = attrs
}

func SetShader(shader *Shader) {
	c := theContext
	c.shader = shader
}

func SetTexture(tex *Texture, slot int) {
	c := theContext
	if c.texture[slot] != tex {
		c.dirtyFlag |= dirtyTexture
		c.texture[slot] = tex
	}
}

func SetTarget(target *Target) {
	c := theContext
	if c.target != target {
		c.dirtyFlag |= dirtyTarget
		c.target = target
	}
}

func SetBlend(src, dst BlendFormat) {
	c := theContext
	c.dirtyFlag |= dirtyBlend
	c.blendSrc = src
	c.blendDst = dst
}

func SetDepth(depth DepthFormat) {
	c := theContext
	c.dirtyFlag |= dirtyDepth
	c.depth = depth
}

func EnableDepthMask(enable bool) {
	c := theContext
	c.dirtyFlag |= dirtyDepth
	c.depthmask = enable
}

func EnableScissor(enable bool) {
	c := theContext
	c.dirtyFlag |= dirtyScissor
	c.scissor = enable
}

func (c *Context) commit() {
	c.shader.bind()

	if c.dirtyFlag&dirtyTexture != 0 {
		for i := 0; i < len(c.shader.samplers); i++ {
			c.texture[i].activeTexture(i)
		}
	}

	if c.dirtyFlag&dirtyTarget != 0 {
		c.target.bind()
	}

	if c.dirtyFlag&dirtyBlend != 0 {
		if c.blendSrc == BlendDisable {
			disable(Blend)
		} else {
			enable(Blend)
			blendFunc(c.blendSrc, c.blendDst)
		}
	}

	if c.dirtyFlag&dirtyDepth != 0 {
		if c.depth == DepthDisable {
			disable(DepthTest)
		} else {
			enable(DepthTest)
			depthFunc(c.depth)
		}
		depthMask(c.depthmask)
	}

	if c.dirtyFlag&dirtyScissor != 0 {
		if c.scissor {
			enable(ScissorTest)
		} else {
			disable(ScissorTest)
		}
	}

	c.dirtyFlag = 0
}

func Draw(mode DrawMode, start, count int) {
	if count > 0 {
		theContext.commit()
		glDraw(mode, start, count)
	}
}
