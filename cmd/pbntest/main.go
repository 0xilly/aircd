package main

import (
	"fmt"
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

	const canvas_width = 240
	const canvas_height = 200
	canvas := libpbn.Create(canvas_width, canvas_height)
	canvas.Clear(nil)

	total_lines := 0
	frame := 0
	base_filename := "out/frame_%04d.png"

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
			canvas.ParseLine(command)
			if total_lines%100 == 0 {
				canvas.SaveToFile(fmt.Sprintf(base_filename, frame))
				frame++
			}
			total_lines++
		}
	}
	canvas.SaveToFile(fmt.Sprintf(base_filename, frame))
}
