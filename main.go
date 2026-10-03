package main

import (
	"bufio"
	"bytes"
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
			if i >= len(data) {
				return nil
			}
			value := data[i]
			i++

			for range length {
				result = append(result, value)
			}
		} else {
			if i+length > len(data) {
				return nil
			}
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

func (sa *SimpleArchiver) DecompressFile(inputPath, outputDir string) error {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inFile.Close()

	r := bufio.NewReader(inFile)

	nameLen, err := r.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read file name length: %w", err)
	}
	if nameLen == 0 {
		return fmt.Errorf("invalid archive: empty file name")
	}

	nameBuf := make([]byte, nameLen)
	if _, err = io.ReadFull(r, nameBuf); err != nil {
		return fmt.Errorf("failed to read file name: %w", err)
	}

	fileName := string(nameBuf)
	if fileName != filepath.Base(fileName) || fileName == "." || fileName == ".." {
		return fmt.Errorf("invalid file name in archive: %q", fileName)
	}

	if err = os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outputPath := filepath.Join(outputDir, fileName)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	w := bufio.NewWriter(outFile)
	defer w.Flush()

	sizeBuf := make([]byte, 2)
	blockBuf := make([]byte, 1<<16)

	for {
		_, err := io.ReadFull(r, sizeBuf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read block size: %w", err)
		}

		blockSize := uint16(sizeBuf[0])<<8 | uint16(sizeBuf[1])
		block := blockBuf[:blockSize]

		if _, err = io.ReadFull(r, block); err != nil {
			return fmt.Errorf("failed to read compressed block: %w", err)
		}

		decompressed := sa.decompress(block)
		if decompressed == nil {
			return fmt.Errorf("failed to decompress block: corrupted data")
		}

		if _, err = w.Write(decompressed); err != nil {
			return fmt.Errorf("failed to write decompressed block: %w", err)
		}
	}

	return nil
}

func bytesWord(n int64) string {
	if n%100 >= 11 && n%100 <= 14 {
		return "байт"
	}
	switch n % 10 {
	case 1:
		return "байт"
	case 2, 3, 4:
		return "байта"
	default:
		return "байт"
	}
}

func fileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
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

	outputDir := "output"

	if err := archiver.DecompressFile(outputPath, outputDir); err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}

	restoredPath := filepath.Join(outputDir, filepath.Base(inputPath))
	fmt.Printf("Файл успешно распакован: %s -> %s\n\n", outputPath, restoredPath)

	original, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}

	restored, err := os.ReadFile(restoredPath)
	if err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}

	compressedSize, err := fileSize(outputPath)
	if err != nil {
		fmt.Println("Ошибка:", err)
		os.Exit(1)
	}

	originalSize := int64(len(original))
	restoredSize := int64(len(restored))

	fmt.Printf("Исходный файл: %s (%d %s)\n", inputPath, originalSize, bytesWord(originalSize))
	fmt.Printf("Сжатый файл: %s (%d %s)\n", outputPath, compressedSize, bytesWord(compressedSize))
	fmt.Printf("Распакованный: %s (%d %s)\n", restoredPath, restoredSize, bytesWord(restoredSize))
	fmt.Printf("Содержимое совпадает: %t\n", bytes.Equal(original, restored))
}
