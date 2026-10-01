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

func (sa *SimpleArchiver) compressEmpty(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	return data
}

func main() {
	archiver := NewArchiver("input.txt")

	r1 := archiver.compressEmpty([]byte{})
	fmt.Printf("Тест 1 (пустой массив): len=%d, data=%v\n", len(r1), r1)

	r2 := archiver.compressEmpty([]byte("A"))
	fmt.Printf("Тест 2 (один символ): len=%d, data=%v\n", len(r2), r2)

	r3 := archiver.compressEmpty([]byte{1, 2, 3})
	fmt.Printf("Тест 3 (несколько символов): len=%d, data=%v\n", len(r3), r3)
}
