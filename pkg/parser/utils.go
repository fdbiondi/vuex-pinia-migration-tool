package parser

import (
	"fmt"
	"runtime"
	"strings"
)

func capitalizeByteSlice(str string) string {
	bs := []byte(str)
	if len(bs) == 0 {
		return ""
	}
	bs[0] = byte(bs[0] - 32)
	return string(bs)
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func removeExtension(path string) string {
	return strings.Split(path, ".")[0]
}

func getFilename(path string) string {
	return strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
}

func insertLine(array []string, index int, value string) []string {
	if len(array) == index {
		return append(array, value)
	}
	array = append(array[:index+1], array[index:]...)
	array[index] = value
	return array
}
