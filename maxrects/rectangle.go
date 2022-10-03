package maxrects

type Rect struct {
	Width  int
	Height int
	X      int
	Y      int
}

func (rect *Rect) Collide(other Rect) bool {
	return (other.X < rect.X+rect.Width &&
		other.X+other.Width > rect.X &&
		other.Y < rect.Y+rect.Height &&
		other.Y+other.Height > rect.Y)
}

func (rect *Rect) Contain(other Rect) bool {
	return (other.X >= rect.X && other.Y >= rect.Y &&
		other.X+other.Width <= rect.X+rect.Width && other.Y+other.Height <= rect.Y+rect.Height)
}
