package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

type FileEntry struct {
	Offset   uint32
	FileName string
}

func UnPak(folder string) (map[string]*[]byte, error) {
	files, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("unable to read input folder")
	}

	result := make(map[string]*[]byte)

	fmt.Printf("UnPAKing Phase\n")
	for _, file := range files {
		if file.IsDir() {
			// skip folders
			continue
		}

		ext := strings.ToUpper(filepath.Ext(file.Name()))
		if ext == ".PAK" {
			r := unpak(filepath.Join(folder, file.Name()))
			maps.Copy(result, r)
		}
	}

	return result, nil
}

func unpak(fileName string) map[string]*[]byte {
	data, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Failed to open %s. (%s)\n", fileName, err)
		os.Exit(1)
	}

	var fileEntries []FileEntry
	var previousOffset uint32

	result := make(map[string]*[]byte)

	for i := 0; i < len(data); {
		var offset uint32
		offset = uint32(data[i]) | uint32(data[i+1])<<8 | uint32(data[i+2])<<16 | uint32(data[i+3])<<24
		i += 4

		if offset >= uint32(len(data)) || offset == 0 || offset <= previousOffset {
			break
		}
		previousOffset = offset

		start := i
		for data[i] != 0 {
			i++
		}
		fileName := string(data[start:i])
		i++

		fileEntries = append(fileEntries, FileEntry{Offset: offset, FileName: fileName})
		fmt.Printf("Extracting %s...\n", fileName)
	}

	for j, fileEntry := range fileEntries {
		var nextOffset uint32
		if j < len(fileEntries)-1 {
			nextOffset = fileEntries[j+1].Offset
		} else {
			nextOffset = uint32(len(data))
		}

		//size := nextOffset - fileEntry.Offset
		content := data[fileEntry.Offset:nextOffset]
		entryName := strings.ToUpper(fileEntry.FileName)
		result[entryName] = &content
	}

	return result
}
