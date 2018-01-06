package gl

type Context struct {
	//context
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
	Uint8 AttrType = 0x1401 //gl.UNSIGNED_BYTE
	Uin16 AttrType = 0x1403 //gl.UNSIGNED_SHORT
	Float AttrType = 0x1406 //gl.FLOAT
)

func (at AttrType) size() int {
	switch at {
	case Uint8:
		return 1
	case Uin16:
		return 2
	case Float:
		return 4
	default:
		panic("size of vertex attribute type: invalid type")
	}
}

type Attr struct {
	Name string
	Num  int
	Type AttrType
}

type Attrs []Attr
