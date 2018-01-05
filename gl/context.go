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
