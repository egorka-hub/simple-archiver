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

func (sa *SimpleArchiver) countRepeating(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}

	buf := make([]byte, 0)
	current := data[0]
	count := 1

	for i := 1; i < len(data); i++ {
		if data[i] == current {
			count++
		} else {
			buf = append(buf, byte(count), current)
			current = data[i]
			count = 1
		}
	}

	buf = append(buf, byte(count), current)

	return buf
}

func (sa *SimpleArchiver) createControlByte(count int, isCompressed bool) byte {
	if count > 127 {
		count = 127
	}

	if isCompressed {
		return byte(128 + count)
	}

	return byte(count)
}

func main() {
	archiver := NewArchiver("input.txt")

	tests := []struct {
		name         string
		count        int
		isCompressed bool
	}{
		{"Сжатая последовательность из 5 символов", 5, true},
		{"Несжатая последовательность из 3 символов", 3, false},
	}

	for i, t := range tests {
		b := archiver.createControlByte(t.count, t.isCompressed)
		fmt.Printf("Тест %d: %s\n", i+1, t.name)
		fmt.Printf("Управляющий байт: %08b\n", b)
		fmt.Printf("Старший бит: %t\n", (b&128) != 0)
		fmt.Printf("Количество: %d\n\n", b&127)
	}

}
