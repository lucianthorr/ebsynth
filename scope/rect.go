package scope

type point struct {
	x, y float32
}

func newPt(x, y float32) *point {
	return &point{
		x: x, y: y,
	}
}

type rect struct {
	// bottom-left, bottom-right, top-left, top-right
	bL, bR, tL, tR *point
}
