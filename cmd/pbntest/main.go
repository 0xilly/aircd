package main

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"strings"

	"github.com/0xilly/aircd/libpbn"
)

func main() {
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
	canvas := libpbn.Create(canvas_width, canvas_height)
	canvas.Clear(nil)

	total_lines := 0
	frame := 0
	base_filename := "out/frame_%07d.png"

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
			if valid {
				total_lines++
				if total_lines < 111360 {
					canvas.Clear(&color.RGBA{255, 255, 255, 255})
					continue
				}
				// if total_lines > 110762 {
				// 	break
				// }
				if total_lines%5 == 0 {
					// if total_lines > 27550 && total_lines < 28000 {
					canvas.SaveToFile(fmt.Sprintf(base_filename, frame))
					frame++
				}
			}
		}
	}
	// fmt.Printf(strconv.Itoa(total_lines))
	// canvas.SaveToFile(fmt.Sprintf(base_filename, total_lines))
}
