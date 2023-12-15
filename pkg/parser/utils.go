package parser

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/template"
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

//go:embed templates/index.tmpl
var indexTmpl embed.FS

func createEntryPoint(filesMap map[string]*os.File) (string, error) {
	key := getFirstKey(filesMap)

	var storeFilename = strings.Replace(filesMap[key].Name(), key, "index", 1)

	pathSlice := strings.Split(storeFilename, "/")
	storeName := pathSlice[len(pathSlice)-2]

	data, err := indexTmpl.ReadFile("templates/index.tmpl")
	if err != nil {
		return "", err
	}

	var values = map[string]string{
		"storeName":          storeName,
		"storeNameTitleCase": cases.Title(language.English).String(storeName),
	}

	err = parseTemplate(string(data), storeFilename, values)
	if err != nil {
		return "", err
	}

	return storeName, nil
}

func parseTemplate(templateStr string, outputPath string, values map[string]string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := template.Must(template.New("template").Parse(templateStr))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, values); err != nil {
		return err
	}

	err = os.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		return errors.New("error wrinting file -> " + err.Error())
	}

	return nil
}

func getFirstKey[K comparable, V any](m map[K]V) K {
	for k := range m {
		return k
	}

	return *new(K)
}
