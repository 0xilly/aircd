package libpbn

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"strconv"
	"strings"
)

type Parser struct {
	Width  int
	Height int
	Data   *image.RGBA
}

func CreateParser(width int, height int) Parser {
	return Parser{
		Width:  width,
		Height: height,
		Data:   image.NewRGBA(image.Rect(0, 0, width, height)),
	}
}

func (p Parser) Clear(fill_color *color.RGBA) {
	if fill_color == nil {
		fill_color = &color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	for x := range p.Width {
		for y := range p.Height {
			p.Data.Set(x, y, fill_color)
		}
	}
}

func (p Parser) ParseLine(line string) bool {
	if len(line) == 0 {
		return false
	}
	if line[0] == '#' {
		return p.parseV1(line)
	}
	line_runes := []rune(line)
	if p.IsColorRune(line_runes, 0) {
		return p.parseV2(line_runes)
	}
	return false
}

func (p Parser) parseV1(line string) bool {
	parts := strings.Split(line, " ")
	if len(parts) != 2 {
		return false
	}
	s_color := parts[0][1:]
	pixel_color := color.RGBA{}
	if len(s_color) == 3 {
		hcolor, err := strconv.ParseUint(s_color, 16, 16)
		if err != nil {
			return false
		}
		pixel_color.R = uint8((hcolor>>8)&0xf) * 16
		pixel_color.G = uint8((hcolor>>4)&0xf) * 16
		pixel_color.B = uint8(hcolor&0xf) * 16
		pixel_color.A = 255

	} else if len(s_color) == 6 {
		hcolor, err := strconv.ParseUint(s_color, 16, 32)
		if err != nil {
			return false
		}
		pixel_color.R = uint8(hcolor >> 16)
		pixel_color.G = uint8(hcolor >> 8)
		pixel_color.B = uint8(hcolor)
		pixel_color.A = 255
	}
	pixels := strings.Split(parts[1], ";")
	for _, pixel := range pixels {
		cord := strings.Split(pixel, ",")
		if len(cord) != 2 {
			return false
		}
		x, err := strconv.ParseInt(cord[0], 10, 16)
		if err != nil {
			return false
		}
		y, err := strconv.ParseInt(cord[1], 10, 16)
		if err != nil {
			return false
		}
		p.Data.Set(int(x), int(y), pixel_color)
	}
	return true
}

func (p Parser) parseV2(line []rune) bool {
	current_color := p.getColor(line, 0)
	index := 2
	for {
		if (index + 1) > len(line) {
			break
		}
		if p.IsColorRune(line, index) {
			// Handle color change
			current_color = p.getColor(line, index)
			index += 2
			continue
		}
		if line[index] >= UNICODE_INDEX_START {
			// Handle pixel index
			pindex := int(line[index] - UNICODE_INDEX_START)
			if pindex >= 0 && pindex < (p.Width*p.Height) {
				p.Data.Set(pindex%p.Width, pindex/p.Width, current_color)
			}
			index += 1
			continue
		}
		if line[index] == UNICODE_RUN_MARK {
			// Handle RLE block
			if index > len(line) {
				break
			}
			run_length := int(line[index+1] - UNICODE_INDEX_START)
			run_start := int(line[index+2] - UNICODE_INDEX_START)
			for i := range run_length {
				p.Data.Set((run_start+i)%p.Width, (run_start+i)/p.Width, current_color)
			}
			index += 2
			continue
		}
		fmt.Println("Invalid rune found while processing", index, len(line))
		return false
	}
	return true
}

func (p Parser) IsColorRune(line []rune, index int) bool {
	return line[index] >= UNICODE_COLOR_START && line[index] <= UNICODE_COLOR_END
}

func (p Parser) getColor(line []rune, index int) color.RGBA {
	rg := line[index]
	ba := line[index+1]
	return color.RGBA{R: uint8(rg & 0xff00 >> 8), G: uint8(rg & 0xff), B: uint8(ba & 0xff00 >> 8), A: uint8(ba & 0xff)}
}

func (p Parser) SaveToFile(filename string) {
	image_file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer image_file.Close()
	err = png.Encode(image_file, p.Data)
	if err != nil {
		log.Fatal(err)
	}
}
