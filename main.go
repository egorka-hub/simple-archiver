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

func main() {
	archiver := NewArchiver("input.txt")
	tests := []string{"AABBB", "ABC"}
	for i, input := range tests {
		r := archiver.countRepeating([]byte(input))

		fmt.Printf("Тест %d (%s):\n", i+1, input)
		fmt.Printf("Вход: %s\n", input)
		fmt.Print("Выход: ")
		for j := 0; j < len(r); j += 2 {
			fmt.Printf("%d%c", r[j], r[j+1])
		}
		fmt.Println()
		fmt.Println()
	}

}
