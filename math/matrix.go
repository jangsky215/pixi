package math

type Matrix struct {
	a, b   float32
	c, d   float32
	tx, ty float32
}

func (m *Matrix) Clone() *Matrix {
	return &Matrix{m.a, m.b, m.c, m.d, m.tx, m.ty}
}

func (m *Matrix) ToArray() []float32 {
	return []float32{
		m.a, m.b, 0,
		m.c, m.d, 0,
		m.tx, m.ty, 1,
	}
}

func (m *Matrix) Apply(x, y float32) (float32, float32) {
	newX := (m.a * x) + (m.c * y) + m.tx
	newY := (m.b * x) + (m.d * y) + m.ty

	return newX, newY
}

func (m *Matrix) ApplyInverse(x, y float32) (float32, float32) {
	id := 1 / ((m.a * m.d) - (m.b * m.c))

	newX := (m.d*x - m.c*y + m.ty*m.c - m.tx*m.d) * id
	newY := (m.a*y - m.b*x - m.ty*m.a + m.tx*m.b) * id

	return newX, newY
}

func (m *Matrix) Identity() {
	*m = Matrix{1, 0, 0, 1, 0, 0}
}

func (m *Matrix) Translate(x, y float32) {
	m.tx += x
	m.ty += y
}

func (m *Matrix) Scale(x, y float32) {
	m.a *= x
	m.c *= x
	m.tx *= x
	m.b *= y
	m.d *= y
	m.ty *= y
}

func (m *Matrix) Rotate(angle float32) {
	if angle == 0 {
		return
	}

	radian := angle * RadianFactor
	cos := Cos(radian)
	sin := Sin(radian)

	a := m.a
	c := m.c
	tx := m.tx

	m.a = (a * cos) - (m.b * sin)
	m.b = (a * sin) + (m.b * cos)
	m.c = (c * cos) - (m.d * sin)
	m.d = (c * sin) + (m.d * cos)
	m.tx = (tx * cos) - (m.ty * sin)
	m.ty = (tx * sin) + (m.ty * cos)
}

func (m *Matrix) Invert() {
	a := m.a
	b := m.b
	c := m.c
	d := m.d
	tx := m.tx
	n := (a * d) - (b * c)

	m.a = d / n
	m.b = -b / n
	m.c = -c / n
	m.d = a / n
	m.tx = (c*m.ty - d*tx) / n
	m.ty = -(a*m.ty - b*m.tx) / n
}

func (m *Matrix) Append(matrix *Matrix) {
	a := m.a
	b := m.b
	c := m.c
	d := m.d

	m.a = (matrix.a * a) + (matrix.b * c)
	m.b = (matrix.a * b) + (matrix.b * d)
	m.c = (matrix.c * a) + (matrix.d * c)
	m.d = (matrix.c * b) + (matrix.d * d)

	m.tx = (matrix.tx * a) + (matrix.ty * c) + m.tx
	m.ty = (matrix.tx * b) + (matrix.ty * d) + m.ty
}
