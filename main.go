package main

import (
	"bytes"
	"fmt"
)

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

func (sa *SimpleArchiver) decompress(data []byte) []byte {
	if len(data) == 0 {
		return []byte{}
	}

	var i int
	var result []byte

	for i < len(data) {
		control := data[i]
		i++

		isCompressed := control&0x80 != 0
		length := int(control & 0x7F)

		if isCompressed {
			value := data[i]
			i++

			for range length {
				result = append(result, value)
			}
		} else {
			result = append(result, data[i:i+length]...)
			i += length
		}
	}
	return result
}

func (sa *SimpleArchiver) testRoundTrip(num int, name string, data []byte) {
	compressed := sa.compress(data)
	decompressed := sa.decompress(compressed)

	fmt.Printf("Тест %d: Сжатие и распаковка %s\n", num, name)
	if len(data) <= 16 {
		fmt.Printf("Исходные данные: %v\n", data)
		fmt.Printf("Сжатые данные:   %v\n", compressed)
		fmt.Printf("Распакованные:   %v\n", decompressed)
	} else {
		fmt.Printf("Исходные данные: %d байт\n", len(data))
		fmt.Printf("Сжатые данные:   %d байт\n", len(compressed))
		fmt.Printf("Распакованные:   %d байт\n", len(decompressed))
	}
	fmt.Printf("Данные совпадают: %t\n\n", bytes.Equal(data, decompressed))
}

func main() {
	archiver := NewArchiver("input.txt")

	unique127 := make([]byte, 127)
	for i := range unique127 {
		unique127[i] = byte(i)
	}

	unique200 := make([]byte, 200)
	for i := range unique200 {
		unique200[i] = byte(i)
	}

	archiver.testRoundTrip(1, "'AAAAABCD'", []byte("AAAAABCD"))
	archiver.testRoundTrip(2, "'ABCDE'", []byte("ABCDE"))
	archiver.testRoundTrip(3, "пустых данных", []byte{})
	archiver.testRoundTrip(4, "'A'", []byte("A"))
	archiver.testRoundTrip(5, "'AAAAABCCCCCCDE'", []byte("AAAAABCCCCCCDE"))
	archiver.testRoundTrip(6, "127 × 'A'", bytes.Repeat([]byte("A"), 127))
	archiver.testRoundTrip(7, "300 × 'A'", bytes.Repeat([]byte("A"), 300))
	archiver.testRoundTrip(8, "127 разных байт", unique127)
	archiver.testRoundTrip(9, "200 разных байт", unique200)
}
