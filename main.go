package main

import "fmt"

type SimpleArchiver struct {
	inputPath  string
	outputPath string
	buffer     []byte
}

func NewArchiver(inputPath string) *SimpleArchiver {
	return &SimpleArchiver{
		inputPath: inputPath,
		buffer:    make([]byte, 1024*8),
	}
}

func main() {
	archiver := NewArchiver("input.txt")
	fmt.Printf("Архиватор создан, размер буфера: %d байт\n", len(archiver.buffer))
	fmt.Printf("путь к исходному файлу: %s\n", archiver.inputPath)
}
