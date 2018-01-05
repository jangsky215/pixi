package gl

type Context struct {
	context
}

var theContext *Context

func GetContext() *Context {
	return theContext
}

type DrawMode int

const (
	DrawTriangle DrawMode = 0x0004 //gl.TRIANGLES
	DrawLine     DrawMode = 0x1B01 //gl.LINE
)

type AttrType int

const (
	Int   AttrType = 0x1404 //gl.INT
	Float AttrType = 0x1406 //gl.FLOAT

	Vec2 AttrType = 0x8B50 //gl.FLOAT_VEC2
	Vec3 AttrType = 0x8B51 //gl.FLOAT_VEC3
	Vec4 AttrType = 0x8B52 //gl.FLOAT_VEC4

	Mat2 AttrType = 0x8B5A //FLOAT_MAT2
	Mat3 AttrType = 0x8B5B //FLOAT_MAT3
	Mat4 AttrType = 0x8B5C //FLOAT_MAT4

	Sampler2D AttrType = 0x8B5E //gl.SAMPLER_2D
)

func (at AttrType) Size() int {
	switch at {
	case Int, Float:
		return 4
	case Vec2:
		return 2 * 4
	case Vec3:
		return 3 * 4
	case Vec4:
		return 4 * 4
	case Mat2:
		return 2 * 2 * 4
	case Mat3:
		return 3 * 3 * 4
	case Mat4:
		return 4 * 4 * 4
	default:
		panic("size of vertex attribute type: invalid type")
	}
}

type Attr struct {
	Name string
	Type AttrType
}
