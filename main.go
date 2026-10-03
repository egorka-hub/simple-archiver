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

func (sa *SimpleArchiver) hasRunAhead(data []byte, i int) bool {
	if i+2 >= len(data) {
		return false
	}
	return data[i] == data[i+1] && data[i+1] == data[i+2]
}

func (sa *SimpleArchiver) compress(data []byte) []byte {
	if len(sa.compressEmpty(data)) == 0 {
		return []byte{}
	}

	result := []byte{}
	i := 0

	for i < len(data) {
		end := min(i+127, len(data))
		n := int(sa.countRepeating(data[i:end])[0])

		if n >= 4 {
			result = append(result, sa.createControlByte(n, true), data[i])
			i += n
		} else {
			group := []byte{data[i]}
			j := i + 1

			for j < len(data) && len(group) < 127 && !sa.hasRunAhead(data, j) {
				group = append(group, data[j])
				j++
			}

			result = append(result, sa.createControlByte(len(group), false))
			result = append(result, group...)
			i = j
		}
	}
	return result
}

func main() {
	archiver := NewArchiver("input.txt")

	tests := []struct {
		name  string
		input string
	}{
		{"Сжатие", "AAAAA"},
		{"Сжатие", "ABBBCCCCDE"},
		{"Сжатие", "ABCDE"},
	}

	for i, t := range tests {
		result := archiver.compress([]byte(t.input))
		fmt.Printf("Тест %d: %s '%s'\n", i+1, t.name, t.input)
		fmt.Printf("Вход: %s\n", t.input)
		fmt.Printf("Результат: % X\n\n", result)
	}
}
