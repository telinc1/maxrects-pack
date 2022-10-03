package maxrects

import (
	"errors"
	"math"
	"sort"
)

type Bin struct {
	Width  int
	Height int

	MaxWidth  int
	MaxHeight int

	Smart  bool
	POT    bool
	Square bool

	Logic int

	Padding int

	FreeRects []Rect

	verticalExpand bool
	stage          Rect
	border         int

	dirty int
}

func NewBin(maxWidth, maxHeight, padding, border int, smart bool) *Bin {
	width, height := 0, 0

	if !smart {
		width = maxWidth
		height = maxHeight
	}

	free := []Rect{
		{
			maxWidth + padding - border*2,
			maxHeight + padding - border*2,
			border,
			border,
		},
	}

	return &Bin{
		Width:          width,
		Height:         height,
		MaxWidth:       maxWidth,
		MaxHeight:      maxHeight,
		Smart:          smart,
		POT:            true,
		Square:         true,
		Logic:          PACK_MAX_EDGE,
		Padding:        padding,
		FreeRects:      free,
		verticalExpand: false,
		stage:          Rect{width, height, 0, 0},
		border:         border,
		dirty:          0,
	}
}

func (bin *Bin) PlaceAll(rects []*Rect) error {
	sort.SliceStable(rects, func(i, k int) bool {
		a, b := rects[i], rects[k]
		result := max(b.Width, b.Height) - max(a.Width, a.Height)
		return result < 0
	})

	for _, rect := range rects {
		if !bin.Place(rect) {
			return errors.New("can't fit all rectangles")
		}
	}

	return nil
}

func (bin *Bin) Place(rect *Rect) bool {
	allowRotation := false
	node := bin.findNode(rect.Width+bin.Padding, rect.Height+bin.Padding, allowRotation)

	if node != nil {
		node := *node

		bin.updateBinSize(node)

		numRectToProcess := len(bin.FreeRects)

		for i := 0; i < numRectToProcess; i++ {
			if bin.splitNode(bin.FreeRects[i], node) {
				bin.FreeRects = append(bin.FreeRects[:i], bin.FreeRects[i+1:]...)

				numRectToProcess--
				i--
			}
		}

		bin.pruneFreeList()

		if bin.Width > bin.Height {
			bin.verticalExpand = true
		} else {
			bin.verticalExpand = false
		}

		rect.X = node.X
		rect.Y = node.Y

		bin.dirty++

		return true
	} else if !bin.verticalExpand {
		if bin.updateBinSize(Rect{
			rect.Width + bin.Padding, rect.Height + bin.Padding,
			bin.Width + bin.Padding - bin.border, bin.border,
		}) || bin.updateBinSize(Rect{
			rect.Width + bin.Padding, rect.Height + bin.Padding,
			bin.border, bin.Height + bin.Padding - bin.border,
		}) {
			return bin.Place(rect)
		}
	} else {
		if bin.updateBinSize(Rect{
			rect.Width + bin.Padding, rect.Height + bin.Padding,
			bin.border, bin.Height + bin.Padding - bin.border,
		}) || bin.updateBinSize(Rect{
			rect.Width + bin.Padding, rect.Height + bin.Padding,
			bin.Width + bin.Padding - bin.border, bin.border,
		}) {
			return bin.Place(rect)
		}
	}

	return false
}

func (bin *Bin) findNode(width, height int, _ bool) *Rect {
	score := math.MaxInt

	var areaFit int
	var bestNode *Rect = nil

	for _, r := range bin.FreeRects {
		if r.Width >= width && r.Height >= height {
			if bin.Logic == PACK_MAX_AREA {
				areaFit = r.Width*r.Height - width*height
			} else {
				areaFit = min(r.Width-width, r.Height-height)
			}

			if areaFit < score {
				bestNode = &Rect{width, height, r.X, r.Y}
				score = areaFit
			}
		}
	}

	return bestNode
}

func (bin *Bin) splitNode(freeRect, usedNode Rect) bool {
	// Test if usedNode intersect with freeRect
	if !freeRect.Collide(usedNode) {
		return false
	}

	// Do vertical split
	if usedNode.X < freeRect.X+freeRect.Width && usedNode.X+usedNode.Width > freeRect.X {
		// New node at the top side of the used node
		if usedNode.Y > freeRect.Y && usedNode.Y < freeRect.Y+freeRect.Height {
			newNode := Rect{freeRect.Width, usedNode.Y - freeRect.Y, freeRect.X, freeRect.Y}

			bin.FreeRects = append(bin.FreeRects, newNode)
		}

		// New node at the bottom side of the used node
		if usedNode.Y+usedNode.Height < freeRect.Y+freeRect.Height {
			newNode := Rect{
				freeRect.Width,
				freeRect.Y + freeRect.Height - (usedNode.Y + usedNode.Height),
				freeRect.X,
				usedNode.Y + usedNode.Height,
			}

			bin.FreeRects = append(bin.FreeRects, newNode)
		}
	}

	// Do Horizontal split
	if usedNode.Y < freeRect.Y+freeRect.Height &&
		usedNode.Y+usedNode.Height > freeRect.Y {
		// New node at the left side of the used node.
		if usedNode.X > freeRect.X && usedNode.X < freeRect.X+freeRect.Width {
			newNode := Rect{usedNode.X - freeRect.X, freeRect.Height, freeRect.X, freeRect.Y}

			bin.FreeRects = append(bin.FreeRects, newNode)
		}

		// New node at the right side of the used node.
		if usedNode.X+usedNode.Width < freeRect.X+freeRect.Width {
			newNode := Rect{
				freeRect.X + freeRect.Width - (usedNode.X + usedNode.Width),
				freeRect.Height,
				usedNode.X + usedNode.Width,
				freeRect.Y,
			}

			bin.FreeRects = append(bin.FreeRects, newNode)
		}
	}

	return true
}

func (bin *Bin) pruneFreeList() {
	// Go through each pair of freeRects and remove any rects that is redundant
	i, j, len := 0, 0, len(bin.FreeRects)

	for i < len {
		j = i + 1

		tmpRect1 := bin.FreeRects[i]

		for j < len {
			tmpRect2 := bin.FreeRects[j]

			if tmpRect2.Contain(tmpRect1) {
				bin.FreeRects = append(bin.FreeRects[:i], bin.FreeRects[i+1:]...)
				i--
				len--

				break
			}

			if tmpRect1.Contain(tmpRect2) {
				bin.FreeRects = append(bin.FreeRects[:j], bin.FreeRects[j+1:]...)
				j--
				len--
			}

			j++
		}

		i++
	}
}

func (bin *Bin) updateBinSize(node Rect) bool {
	if !bin.Smart {
		return false
	}

	if bin.stage.Contain(node) {
		return false
	}

	tmpWidth := max(bin.Width, node.X+node.Width-bin.Padding+bin.border)
	tmpHeight := max(bin.Height, node.Y+node.Height-bin.Padding+bin.border)

	if bin.POT {
		tmpWidth = int(math.Pow(2, math.Ceil(math.Log2(float64(tmpWidth)))))
		tmpHeight = int(math.Pow(2, math.Ceil(math.Log2(float64(tmpHeight)))))
	}

	if bin.Square {
		tmpWidth = max(tmpWidth, tmpHeight)
		tmpHeight = tmpWidth
	}

	if tmpWidth > bin.MaxWidth+bin.Padding || tmpHeight > bin.MaxHeight+bin.Padding {
		return false
	}

	bin.expandFreeRects(tmpWidth+bin.Padding, tmpHeight+bin.Padding)

	bin.stage.Width = tmpWidth
	bin.stage.Height = tmpHeight

	bin.Width = tmpWidth
	bin.Height = tmpHeight

	return true
}

func (bin *Bin) expandFreeRects(width, height int) {
	for i := range bin.FreeRects {
		freeRect := &bin.FreeRects[i]

		if freeRect.X+freeRect.Width >= min(bin.Width+bin.Padding-bin.border, width) {
			freeRect.Width = width - freeRect.X - bin.border
		}

		if freeRect.Y+freeRect.Height >= min(bin.Height+bin.Padding-bin.border, height) {
			freeRect.Height = height - freeRect.Y - bin.border
		}
	}

	bin.FreeRects = append(
		bin.FreeRects,
		Rect{width - bin.Width - bin.Padding,
			height - bin.border*2,
			bin.Width + bin.Padding - bin.border,
			bin.border},
		Rect{width - bin.border*2,
			height - bin.Height - bin.Padding,
			bin.border,
			bin.Height + bin.Padding - bin.border},
	)

	newFreeRects := make([]Rect, 0, len(bin.FreeRects))

	for i := range bin.FreeRects {
		freeRect := bin.FreeRects[i]

		if !(freeRect.Width <= 0 || freeRect.Height <= 0 || freeRect.X < bin.border || freeRect.Y < bin.border) {
			newFreeRects = append(newFreeRects, freeRect)
		}
	}

	bin.FreeRects = newFreeRects

	bin.pruneFreeList()
}
