package libpbn

import (
	"fmt"
	"image"
	"image/color"
)

type Encoder struct {
	Width     int
	Height    int
	Data      *image.RGBA
	MaxLength int
	lines     []string
	line      []rune
	lastColor *color.RGBA
}

func CreateEncoder(width int, height int, max_length int) Encoder {
	return Encoder{
		Width:     width,
		Height:    height,
		Data:      image.NewRGBA(image.Rect(0, 0, width, height)),
		MaxLength: max_length,
		lines:     make([]string, 0),
		line:      make([]rune, 0),
		lastColor: &color.RGBA{255, 255, 255, 255},
	}
}

func (e *Encoder) Clear(fill_color *color.RGBA) {
	if fill_color == nil {
		fill_color = &color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	for x := range e.Width {
		for y := range e.Height {
			e.Data.Set(x, y, fill_color)
		}
	}
}

func (e *Encoder) EncodeFrameV1(new_frame *image.RGBA, max_length *int) []string {
	lines := make([]string, 0)
	colors := make(map[color.RGBA][]string)
	for y := range e.Height {
		for x := range e.Width {
			if e.Data.At(x, y) == new_frame.At(x, y) {
				continue
			}
			if colors[new_frame.RGBAAt(x, y)] == nil {
				colors[new_frame.RGBAAt(x, y)] = make([]string, 0)
			}
			colors[new_frame.RGBAAt(x, y)] = append(colors[new_frame.RGBAAt(x, y)], encodeCoordinates(x, y))
		}
	}
	for n_color := range colors {
		str_color := encodeHex(&n_color)
		line := str_color
		for _, coordinates := range colors[n_color] {
			if max_length != nil && len(line+coordinates) > *max_length {
				lines = append(lines, line)
				line = str_color
			}
			line = line + coordinates
		}
		if line != str_color {
			lines = append(lines, line)
		}
	}
	return lines
}

func encodeHex(i_color *color.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x ", i_color.R, i_color.G, i_color.B)
}
func encodeCoordinates(x int, y int) string {
	return fmt.Sprintf("%d,%d;", x, y)
}

func (e *Encoder) EncodeFrameV2(new_frame *image.RGBA, max_length *int) []string {
	colors := make(map[color.RGBA][]int)
	for y := range e.Height {
		for x := range e.Width {
			if e.Data.At(x, y) == new_frame.At(x, y) {
				continue
			}
			if colors[new_frame.RGBAAt(x, y)] == nil {
				colors[new_frame.RGBAAt(x, y)] = make([]int, 0)
			}
			colors[new_frame.RGBAAt(x, y)] = append(colors[new_frame.RGBAAt(x, y)], e.calculateIndex(x, y))
		}
	}
	e.flushLines()
	for n_color, pixels := range colors {
		index := 0
		for index < len(pixels) {
			start := pixels[index]
			index++
			if index == len(pixels) {
				e.addIndex(&n_color, start)
				continue
			}
			end := pixels[index]
			if end != start+1 {
				e.addIndex(&n_color, start)
				continue
			}
			for ; index < len(pixels) && pixels[index]-pixels[index-1] == 1; index++ {
			}
			end = pixels[index-1]
			if end-start > 1 {
				e.addRun(&n_color, start, end-start+1)
			} else {
				e.addIndex(&n_color, start)
				e.addIndex(&n_color, end)
			}
		}
	}
	e.flushLine(e.lastColor)
	return e.lines
}

func (e *Encoder) calculateIndex(x int, y int) int {
	return (e.Width * y) + x
}

func encodeIndex(index int) rune {
	return UNICODE_INDEX_START + rune(index)
}

func encodeColor(i_color *color.RGBA) []rune {
	return []rune{
		(UNICODE_COLOR_START + (int32(i_color.R) << 8) + int32(i_color.G)),
		(UNICODE_COLOR_START + (int32(i_color.B) << 8) + int32(i_color.A)),
	}
}

func encodeRun(start int, length int) []rune {
	return []rune{
		UNICODE_RUN_MARK,
		UNICODE_INDEX_START + rune(length),
		UNICODE_INDEX_START + rune(start),
	}
}

func (e *Encoder) addIndex(i_color *color.RGBA, index int) {
	if i_color == e.lastColor {
		if len(e.line)+RUNES_PER_INDEX > e.MaxLength {
			e.flushLine(i_color)
		}
		e.line = append(e.line, encodeIndex(index))
	} else {
		if len(e.line)+RUNES_PER_INDEX+RUNES_PER_COLOR > e.MaxLength {
			e.flushLine(i_color)
		} else {
			e.line = append(e.line, encodeColor(i_color)...)
		}
		e.line = append(e.line, encodeIndex(index))
	}
	e.lastColor = i_color
}

func (e *Encoder) addRun(i_color *color.RGBA, start int, length int) {
	if i_color == e.lastColor {
		if len(e.line)+RUNES_PER_RUN > e.MaxLength {
			e.flushLine(i_color)
		}
		e.line = append(e.line, encodeRun(start, length)...)
	} else {
		if len(e.line)+RUNES_PER_RUN+RUNES_PER_COLOR > e.MaxLength {
			e.flushLine(i_color)
		} else {
			e.line = append(e.line, encodeColor(i_color)...)
		}
		e.line = append(e.line, encodeRun(start, length)...)
	}
	e.lastColor = i_color
}

func (e *Encoder) flushLine(i_color *color.RGBA) {
	if len(e.line) > RUNES_PER_COLOR {
		e.lines = append(e.lines, string(e.line))
	}
	e.line = encodeColor(i_color)
}

func (e *Encoder) flushLines() {
	e.lines = make([]string, 0)
	e.line = make([]rune, 0)
}
