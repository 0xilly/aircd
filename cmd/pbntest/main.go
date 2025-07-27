package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/0xilly/aircd/libpbn"
)

func parse_test() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	buffer := make([]byte, 4096)
	tail := ""

	// 0-27565 123x92
	// 27565-101200 132x99
	// 105700-110762 192x144
	// 111360-end 160x120

	// const canvas_width = 240
	// const canvas_height = 200
	const canvas_width = 160
	const canvas_height = 120
	// const canvas_width = 132
	// const canvas_height = 99
	// const canvas_width = 123
	// const canvas_height = 92
	// const canvas_width = 192
	// const canvas_height = 144
	canvas := libpbn.CreateParser(canvas_width, canvas_height)
	canvas.Clear(nil)

	total_lines := 0
	valid_lines := 0

	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalln(err)
		}
		current := tail + string(buffer[:n])
		lines := strings.Split(current, "\n")
		tail = lines[len(lines)-1]
		if len(tail) > 0 {
			lines = lines[0 : len(lines)-1]
		}
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
			command := strings.Split(line, "\t")[2]
			valid := canvas.ParseLine(command)
			total_lines++
			if valid {
				valid_lines++
			}
		}
	}
	fmt.Printf("Total: %d\nValid: %d\n", total_lines, valid_lines)
}

func encode_test() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
	img, err := png.Decode(file)
	if err != nil {
		log.Fatalln(err)
	}

	bounds := img.Bounds()
	rgbaImage := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(rgbaImage, rgbaImage.Bounds(), img, bounds.Min, draw.Src)

	max_length := 1000000

	canvas_width := bounds.Dx()
	canvas_height := bounds.Dy()
	encoder := libpbn.CreateEncoder(canvas_width, canvas_height, max_length)
	parser := libpbn.CreateParser(canvas_width, canvas_height)

	commands := make([]string, 0)
	start_time := time.Now()

	encoder.Clear(nil)
	parser.Clear(nil)
	start_time = time.Now()
	commands = encoder.EncodeFrameV1(rgbaImage, &max_length)
	fmt.Printf("V1 encoded in %d lines\n", len(commands))
	fmt.Printf("encoding took %s\n", time.Since(start_time).String())
	for _, command := range commands {
		parser.ParseLine(command)
	}
	parser.SaveToFile("v1.png")

	encoder.Clear(nil)
	parser.Clear(nil)
	start_time = time.Now()
	commands = encoder.EncodeFrameV2(rgbaImage, &max_length)
	fmt.Printf("V2 encoded in %d lines\n", len(commands))
	fmt.Printf("encoding took %s\n", time.Since(start_time).String())
	valid_count := 0
	for _, command := range commands {
		fmt.Println(len([]rune(command)))
		valid := parser.ParseLine(command)
		if valid {
			valid_count++
		}
	}
	fmt.Printf("V2 decoded %d valid lines\n", valid_count)
	parser.SaveToFile("v2.png")
}

func main() {
	// parse_test()
	encode_test()
}
