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
	Data   []uint8
}

func Create(width int, height int) Parser {
	return Parser{
		Width:  width,
		Height: height,
		Data:   make([]uint8, width*height*3),
	}
}

func (p Parser) Clear(fill_color *color.RGBA) {
	if fill_color == nil {
		fill_color = &color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	for index := 0; index < len(p.Data); index += 3 {
		p.Data[index] = fill_color.R
		p.Data[index+1] = fill_color.G
		p.Data[index+2] = fill_color.B
	}
}

func (p Parser) ParseLine(line string) {
	if line[0] == '#' {
		p.parseV1(line)
	}
}

func (p Parser) parseV1(line string) {
	parts := strings.Split(line, " ")
	if len(parts) != 2 {
		return
	}
	s_color := parts[0][1:]
	pixel_color := color.RGBA{}
	if len(s_color) == 3 {
		hcolor, err := strconv.ParseUint(s_color, 16, 16)
		if err != nil {
			return
		}
		pixel_color.R = uint8((hcolor>>8)&0xf) * 16
		pixel_color.G = uint8((hcolor>>4)&0xf) * 16
		pixel_color.B = uint8(hcolor&0xf) * 16
		pixel_color.A = 255

	} else if len(s_color) == 6 {
		hcolor, err := strconv.ParseUint(s_color, 16, 32)
		if err != nil {
			return
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
			return
		}
		x, err := strconv.ParseInt(cord[0], 10, 16)
		if err != nil {
			return
		}
		y, err := strconv.ParseInt(cord[1], 10, 16)
		if err != nil {
			return
		}
		p.SetPixel(int(x), int(y), pixel_color)
	}
}

func (p Parser) GetIndex(x int, y int) int {
	return (x + (p.Width * y)) * 3
}

func (p Parser) GetPixel(x int, y int) color.RGBA {
	index := p.GetIndex(x, y)
	return color.RGBA{R: p.Data[index], G: p.Data[index+1], B: p.Data[index+2], A: 255}
}
func (p Parser) SetPixel(x int, y int, color color.RGBA) {
	if x >= p.Width || y >= p.Height || x < 0 || y < 0 {
		return
	}
	index := p.GetIndex(x, y)
	p.Data[index] = color.R
	p.Data[index+1] = color.G
	p.Data[index+2] = color.B
}

func (p Parser) Print() {
	for y := 0; y < p.Height; y++ {
		line := ""
		for x := 0; x < p.Width; x++ {
			pixel := p.GetPixel(x, y)
			line += fmt.Sprintf("%02x%02x%02x ", pixel.R, pixel.G, pixel.B)
		}
		fmt.Println(line)
	}
}

func (p Parser) SaveToFile(filename string) {
	img := image.NewRGBA(image.Rect(0, 0, p.Width, p.Height))
	for y := range p.Height {
		for x := range p.Width {
			pixel := p.GetPixel(x, y)
			img.SetRGBA(x, y, pixel)
		}
	}
	image_file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer image_file.Close()
	err = png.Encode(image_file, img)
	if err != nil {
		log.Fatal(err)
	}
}
