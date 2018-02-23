package internal

import "math"

type AttrType int

const (
	Uin16 AttrType = 0x1403 //gl.UNSIGNED_SHORT
	Float AttrType = 0x1406 //gl.FLOAT
)

func (at AttrType) size() int {
	switch at {
	case Uin16:
		return 2
	case Float:
		return 4
	default:
		panic("size of vertex attribute type: invalid type")
	}
}

func (at AttrType) normalized() bool {
	return at != Float
}

type Attr struct {
	Name string
	Num  int
	Type AttrType
}

type Attrs []Attr

type BlendFormat uint32

const (
	BlendDisable          BlendFormat = math.MaxUint32
	BlendZero             BlendFormat = 0      //gl.ZERO
	BlendOne              BlendFormat = 1      //gl.ONE
	BlendSrcColor         BlendFormat = 0x0300 //gl.SRC_COLOR
	BlendOneMinusSrcColor BlendFormat = 0x0301 //gl.ONE_MINUS_SRC_COLOR
	BlendSrcAlpha         BlendFormat = 0x0302 //gl.SRC_ALPHA
	BlendOneMinusSrcAlpha BlendFormat = 0x0303 //gl.ONE_MINUS_SRC_ALPHA
	BlendDstAlpha         BlendFormat = 0x0304 //gl.DST_ALPHA
	BlendOneMinusDstAlpha BlendFormat = 0x0305 //gl.ONE_MINUS_DST_ALPHA
	BlendDstColor         BlendFormat = 0x0306 //gl.DST_COLOR
	BlendOneMinusDstColor BlendFormat = 0x0307 //gl.ONE_MINUS_DST_COLOR
	BlendSrcAlphaSaturate BlendFormat = 0x0308 //gl.SRC_ALPHA_SATURATE
)

type DepthFormat uint32

const (
	DepthDisable      DepthFormat = math.MaxUint32
	DepthLess         DepthFormat = 0x0201 //gl.LESS
	DepthEqual        DepthFormat = 0x0202 //gl.EQUAL
	DepthLessEqual    DepthFormat = 0x0203 //gl.LEQUAL
	DepthGreater      DepthFormat = 0x0204 //gl.GREATER
	DepthGreaterEqual DepthFormat = 0x0206 //gl.GEQUAL
	DepthAlways       DepthFormat = 0x0207 //gl.ALWAYS
)

type CapType uint32

const (
	Blend       CapType = 0x0BE2 //gl.BLEND
	DepthTest   CapType = 0x0B71 //gl.DEPTH_TEST
	ScissorTest CapType = 0x0C11 //gl.SCISSOR_TEST
)
