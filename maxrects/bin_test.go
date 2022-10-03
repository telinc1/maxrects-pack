package maxrects

import (
	"math/rand"
	"testing"
	"time"
)

const (
	TEST_PADDING = 4
	TEST_BORDER  = 5
	TEST_REPEAT  = 5
)

func newTestBin(padding, border int) *Bin {
	bin := NewBin(1024, 1024, padding, border, true)
	bin.POT = true
	bin.Square = false

	return bin
}

func doMonkeyTest(t *testing.T, padding, border int, r *rand.Rand) {
	bin := newTestBin(padding, border)

	rects := make([]*Rect, 0, 50)

	for {
		rect := &Rect{r.Intn(200), r.Intn(200), 0, 0}

		if bin.Place(rect) {
			rects = append(rects, rect)
		} else {
			break
		}
	}

	if bin.Width > 1024 {
		t.Fatalf("bin.Width = %v, want at most 1024", bin.Width)
	}

	if bin.Height > 1024 {
		t.Fatalf("bin.Height = %v, want at most 1024", bin.Height)
	}

	for _, rect1 := range rects {
		// Make sure rects are not overlapping
		for _, rect2 := range rects {
			if rect1 != rect2 && rect1.Collide(*rect2) {
				t.Fatalf("intersection detected: %v %v", rect1, rect2)
			}
		}

		// Make sure no rect is outside bounds
		if rect1.X < 0 || rect1.Y < 0 || rect1.X+rect1.Width > bin.Width || rect1.Y+rect1.Height > bin.Height {
			t.Fatalf("rect out of bounds: %v", rect1)
		}
	}
}

func TestNoPaddingIsInitiallyEmpt(t *testing.T) {
	bin := newTestBin(0, 0)

	if bin.Width != 0 {
		t.Fatalf("bin.Width = %v, want 0", bin.Width)
	}

	if bin.Height != 0 {
		t.Fatalf("bin.Height = %v, want 0", bin.Height)
	}
}

func TestNoPaddingAddsRectsCorrectly(t *testing.T) {
	bin := newTestBin(0, 0)
	rect := &Rect{200, 100, 0, 0}

	if !bin.Place(rect) {
		t.Fatalf("bin.Place() failed")
	}

	if rect.X != 0 {
		t.Fatalf("rect.X = %v, want 0", rect.X)
	}

	if rect.Y != 0 {
		t.Fatalf("rect.Y = %v, want 0", rect.Y)
	}
}

func TestNoPaddingUpdatesSizeCorrectly(t *testing.T) {
	bin := newTestBin(0, 0)
	rect := &Rect{200, 100, 0, 0}

	if !bin.Place(rect) {
		t.Fatalf("bin.Place() failed")
	}

	if bin.Width != 256 {
		t.Fatalf("bin.Width = %v, want 256", bin.Width)
	}

	if bin.Height != 128 {
		t.Fatalf("bin.Height = %v, want 128", bin.Height)
	}
}

func TestNoPaddingFitsSquaresCorrectly(t *testing.T) {
	bin := newTestBin(0, 0)

	for i := 0; i < 100; i++ {
		rect := &Rect{100, 100, 0, 0}

		if !bin.Place(rect) {
			t.Fatalf("bin.Place() failed")
		}
	}

	if bin.Width != 1024 {
		t.Fatalf("bin.Width = %v, want 1024", bin.Width)
	}

	if bin.Height != 1024 {
		t.Fatalf("bin.Height = %v, want 1024", bin.Height)
	}
}

func TestNoPaddingMonkeyTesting(t *testing.T) {
	doMonkeyTest(t, 0, 0, rand.New(rand.NewSource(time.Now().UnixNano())))
}

func TestWithPaddingIsInitiallyEmpt(t *testing.T) {
	bin := newTestBin(TEST_PADDING, 0)

	if bin.Width != 0 {
		t.Fatalf("bin.Width = %v, want 0", bin.Width)
	}

	if bin.Height != 0 {
		t.Fatalf("bin.Height = %v, want 0", bin.Height)
	}
}

func TestWithPaddingHandlesPaddingCorrectly(t *testing.T) {
	bin := newTestBin(TEST_PADDING, 0)

	if !bin.Place(&Rect{512, 512, 0, 0}) ||
		!bin.Place(&Rect{512 - TEST_PADDING, 512, 0, 0}) ||
		!bin.Place(&Rect{512, 512 - TEST_PADDING, 0, 0}) {
		t.Fatalf("bin.Place() failed")
	}

	if bin.Width != 1024 {
		t.Fatalf("bin.Width = %v, want 1024", bin.Width)
	}

	if bin.Height != 1024 {
		t.Fatalf("bin.Height = %v, want 1024", bin.Height)
	}
}

func TestWithPaddingAddsRectsWithSizesCloseToTheMax(t *testing.T) {
	bin := newTestBin(TEST_PADDING, 0)

	if !bin.Place(&Rect{1024, 1024, 0, 0}) {
		t.Fatalf("bin.Place() failed")
	}
}

func TestWithPaddingMonkeyTesting(t *testing.T) {
	doMonkeyTest(t, TEST_PADDING, 0, rand.New(rand.NewSource(time.Now().UnixNano())))
}

func TestWithBorderIsInitiallyEmpt(t *testing.T) {
	bin := newTestBin(TEST_PADDING, TEST_BORDER)

	if bin.Width != 0 {
		t.Fatalf("bin.Width = %v, want 0", bin.Width)
	}

	if bin.Height != 0 {
		t.Fatalf("bin.Height = %v, want 0", bin.Height)
	}
}

func TestWithBorderHandlesBorderAndPaddingCorrectly(t *testing.T) {
	bin := newTestBin(TEST_PADDING, TEST_BORDER)

	size := 512 - TEST_BORDER*2

	rect1 := &Rect{size + 1, size, 0, 0}

	if !bin.Place(rect1) {
		t.Fatalf("bin.Place() failed")
	}

	if rect1.X != 5 {
		t.Fatalf("rect1.X = %v, want 5", rect1.X)
	}

	if rect1.Y != 5 {
		t.Fatalf("rect1.Y = %v, want 5", rect1.Y)
	}

	if bin.Width != 1024 {
		t.Fatalf("bin.Width = %v, want 1024", bin.Width)
	}

	if bin.Height != 512 {
		t.Fatalf("bin.Height = %v, want 512", bin.Height)
	}

	rect2 := &Rect{size, size, 0, 0}

	if !bin.Place(rect2) {
		t.Fatalf("bin.Place() failed")
	}

	// Handle space correctly
	if rect2.X-rect1.X-rect1.Width != TEST_PADDING {
		t.Fatalf("incorrect spacing (got %v, want %v): %v %v", rect2.X-rect1.X-rect1.Width, TEST_PADDING, rect1, rect2)
	}

	if rect2.Y != TEST_BORDER {
		t.Fatalf("rect2.Y = %v, want %v", rect1.Y, TEST_BORDER)
	}

	if bin.Width != 1024 {
		t.Fatalf("bin.Width = %v, want 1024", bin.Width)
	}

	if bin.Height != 512 {
		t.Fatalf("bin.Height = %v, want 512", bin.Height)
	}

	if !bin.Place(&Rect{size, size, 0, 0}) {
		t.Fatalf("bin.Place() failed")
	}

	if bin.Place(&Rect{512, 508, 0, 0}) {
		t.Fatalf("bin.Place() didn't fail")
	}

	if bin.Width != 1024 {
		t.Fatalf("bin.Width = %v, want 1024", bin.Width)
	}

	if bin.Height != 1024 {
		t.Fatalf("bin.Height = %v, want 512", bin.Height)
	}
}

func TestWithBorderAddsRectsWithSizesCloseToTheMax(t *testing.T) {
	bin := newTestBin(TEST_PADDING, TEST_BORDER)

	if bin.Place(&Rect{1024, 1024, 0, 0}) {
		t.Fatalf("bin.Place() didn't fail")
	}
}

func TestWithBorderSuperMonkeyTesting(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < TEST_REPEAT; i++ {
		doMonkeyTest(t, r.Intn(10), r.Intn(20), r)
	}
}
