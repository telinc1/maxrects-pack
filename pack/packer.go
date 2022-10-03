package pack

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"image"
	"image/draw"
	"image/png"

	"github.com/telinc1/maxrects-pack/maxrects"
)

type FrameData struct {
	Name string
	Img  image.Image
	Rect maxrects.Rect
}

type OutputMode int8

const (
	OUTPUT_OPTIMAL OutputMode = iota
	OUTPUT_FAST
)

type SingleBinPacker struct {
	MaxWidth  int
	MaxHeight int

	POT    bool
	Square bool

	Extrude int

	Mode OutputMode
}

func NewSingleBinPacker() *SingleBinPacker {
	return &SingleBinPacker{
		MaxWidth:  4096,
		MaxHeight: 4096,
		POT:       true,
		Square:    false,
		Extrude:   0,
		Mode:      OUTPUT_OPTIMAL,
	}
}

func FrameDataFromPNG(name string, reader io.Reader) (*FrameData, error) {
	img, format, err := image.Decode(reader)

	if err != nil {
		return nil, err
	}

	if format != "png" {
		return nil, fmt.Errorf("expected PNG image (got %v)", format)
	}

	// PNGs should always be straight alpha
	if _, ok := img.(*image.NRGBA); !ok {
		return nil, errors.New("image is not NRGBA (expected PNG)")
	}

	return &FrameData{
		name,
		img,
		maxrects.Rect{Width: img.Bounds().Dx(), Height: img.Bounds().Dy(), X: 0, Y: 0},
	}, nil
}

func DrawExtruded(dst draw.Image, dp image.Point, size image.Point, src image.Image, sp image.Point, extrude int) {
	if extrude > 1 {
		panic("extrude > 1 not yet implemented")
	}

	dest := func(x, y, width, height int) image.Rectangle {
		topLeft := dp.Add(image.Point{x - 1, y - 1})
		return image.Rectangle{topLeft, topLeft.Add(image.Point{width, height})}
	}

	// Copy the full source frame.
	draw.Draw(dst, dest(1, 1, size.X, size.Y), src, sp, draw.Src)

	if extrude > 0 {
		// Extrude the top, bottom, left, and right row.
		draw.Draw(dst, dest(1, 0, size.X, 1), src, sp, draw.Src)
		draw.Draw(dst, dest(1, size.Y+1, size.X, 1), src, sp.Add(image.Point{0, size.Y - 1}), draw.Src)
		draw.Draw(dst, dest(0, 1, 1, size.Y), src, sp, draw.Src)
		draw.Draw(dst, dest(size.X+1, 1, 1, size.Y), src, sp.Add(image.Point{size.X - 1, 0}), draw.Src)

		// Extrude the corners (TL, TR, BL, BR).
		dst.Set(dp.X-1, dp.Y-1, src.At(sp.X, sp.Y))
		dst.Set(dp.X+size.X, dp.Y-1, src.At(sp.X+size.X-1, sp.Y))
		dst.Set(dp.X-1, dp.Y+size.Y, src.At(sp.X, sp.Y+size.Y-1))
		dst.Set(dp.X+size.X, dp.Y+size.Y, src.At(sp.X+size.X-1, sp.Y+size.Y-1))
	}
}

func (packer *SingleBinPacker) NewBin() *maxrects.Bin {
	bin := maxrects.NewBin(packer.MaxWidth, packer.MaxHeight, 2*packer.Extrude, packer.Extrude, true)
	bin.POT = packer.POT
	bin.Square = packer.Square

	return bin
}

func (packer *SingleBinPacker) DrawFrame(dst draw.Image, frame *FrameData) {
	DrawExtruded(
		dst, image.Point{frame.Rect.X, frame.Rect.Y}, image.Point{frame.Img.Bounds().Dx(), frame.Img.Bounds().Dy()},
		frame.Img, image.Point{},
		packer.Extrude,
	)
}

// Pack frames into a spritesheet.
// Frame names and frame sources should have the same length and order.
//
// Panics if frameNames doesn't match up with frameSrcs.
func (packer *SingleBinPacker) Pack(
	frameNames []string,
	frameSrcs []io.Reader,
	metadata map[string]any,
	outImage io.Writer,
	outJSON io.Writer,
) error {
	if len(frameNames) != len(frameSrcs) {
		panic("mismatch between frameNames and frameSrcs")
	}

	var group sync.WaitGroup
	var errs errorGroup

	frames := make([]*FrameData, len(frameSrcs))
	rects := make([]*maxrects.Rect, len(frameSrcs))

	for i, frameName := range frameNames {
		reader := frameSrcs[i]

		group.Add(1)

		go func(index int, frameName string, reader io.Reader) {
			defer group.Done()

			data, err := FrameDataFromPNG(frameName, reader)

			if err != nil {
				errs.Add(err)
				return
			}

			frames[index] = data
			rects[index] = &data.Rect
		}(i, frameName, reader)
	}

	group.Wait()

	if !errs.Empty() {
		return errs.Collect()
	}

	bin := packer.NewBin()

	if err := bin.PlaceAll(rects); err != nil {
		return err
	}

	spritesheetJSON := BuildJSON(frames, metadata)

	result := image.NewNRGBA(image.Rectangle{image.Point{}, image.Point{bin.Width, bin.Height}})

	for _, frame := range frames {
		group.Add(1)

		go func(dst *image.NRGBA, frame *FrameData) {
			defer group.Done()

			packer.DrawFrame(dst, frame)
		}(result, frame)
	}

	group.Wait()

	var pngEncoder png.Encoder

	if packer.Mode == OUTPUT_FAST {
		pngEncoder.CompressionLevel = png.BestSpeed
	} else {
		pngEncoder.CompressionLevel = png.BestCompression
	}

	if err := pngEncoder.Encode(outImage, result); err != nil {
		return err
	}

	jsonEncoder := json.NewEncoder(outJSON)

	if packer.Mode == OUTPUT_FAST {
		jsonEncoder.SetIndent("", "    ")
	}

	if err := jsonEncoder.Encode(spritesheetJSON); err != nil {
		return err
	}

	return nil
}
