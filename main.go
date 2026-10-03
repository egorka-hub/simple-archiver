package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

func (sa *SimpleArchiver) CompressFile(inputPath, outputPath string) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inFile.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	fileName := filepath.Base(inputPath)
	if len(fileName) > 255 {
		return fmt.Errorf("file name is too long (%d bytes, max 255): %s", len(fileName), fileName)
	}

	r := bufio.NewReader(inFile)
	w := bufio.NewWriter(outFile)
	defer w.Flush()

	if err = w.WriteByte(byte(len(fileName))); err != nil {
		return fmt.Errorf("failed to write file name length: %w", err)
	}

	if _, err = w.WriteString(fileName); err != nil {
		return fmt.Errorf("failed to write file name: %w", err)
	}

	for {
		n, err := r.Read(sa.buffer)
		if n > 0 {
			compressed := sa.compress(sa.buffer[:n])
			blockSize := len(compressed)

			if err := w.WriteByte(byte(blockSize >> 8)); err != nil {
				return fmt.Errorf("failed to write block size: %w", err)
			}
			if err := w.WriteByte(byte(blockSize)); err != nil {
				return fmt.Errorf("failed to write block size: %w", err)
			}

			if _, err := w.Write(compressed); err != nil {
				return fmt.Errorf("failed to write compressed block: %w", err)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	}

	return nil
}

func main() {
	archiver := NewArchiver("input.txt")

	inputPath := "test.txt"
	outputPath := "test.sarch"

	if err := archiver.CompressFile(inputPath, outputPath); err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}

	fmt.Printf("Файл успешно сжат: %s -> %s\n", inputPath, outputPath)

	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}

	outputInfo, err := os.Stat(outputPath)
	if err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}

	fmt.Printf("Размер исходного файла: %d байт\n", inputInfo.Size())
	fmt.Printf("Размер сжатого файла: %d байт\n", outputInfo.Size())
}
