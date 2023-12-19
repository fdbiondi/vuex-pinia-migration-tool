package parser

import (
	"fmt"
	"reflect"
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

func kebabToCamelCase(kebab string, firstToUpper ...bool) (camelCase string) {
	isToUpper := len(firstToUpper) > 0 && reflect.TypeOf(firstToUpper[0]).Kind() == reflect.Bool && firstToUpper[0]

	for _, runeValue := range kebab {
		if isToUpper {
			camelCase += strings.ToUpper(string(runeValue))
			isToUpper = false
		} else {
			if runeValue == '-' {
				isToUpper = true
			} else {
				camelCase += string(runeValue)
			}
		}
	}
	return
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

func printOutput(path string, fn func()) {
	tag := fmt.Sprintf("--------------------%s--------------------", strings.Split(path, "/")[len(strings.Split(path, "/"))-2])

	if Verbose {
		fmt.Println(tag)
	}

	fn()

	if Verbose {
		for range tag {
			fmt.Printf("-")
		}

		fmt.Println()
		fmt.Println()
	}
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
