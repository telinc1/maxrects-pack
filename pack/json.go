package pack

type RectJSON struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"w"`
	Height int `json:"h"`
}

type FrameJSON struct {
	Frame RectJSON `json:"frame"`
}

type SpritesheetJSON struct {
	Frames map[string]FrameJSON   `json:"frames"`
	Meta   map[string]interface{} `json:"meta"`
}

func BuildJSON(frames []*FrameData, metadata map[string]interface{}) *SpritesheetJSON {
	spritesheetJSON := new(SpritesheetJSON)

	spritesheetJSON.Frames = make(map[string]FrameJSON)
	spritesheetJSON.Meta = metadata

	for _, frame := range frames {
		spritesheetJSON.Frames[frame.Name] = FrameJSON{
			RectJSON{
				frame.Rect.X,
				frame.Rect.Y,
				frame.Img.Bounds().Dx(),
				frame.Img.Bounds().Dy(),
			},
		}
	}

	return spritesheetJSON
}
