package tool

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func Transform(destDir string) error {
	currentPath := ""
	filesToProcess := []string{}

	PrintMemUsage()

	err := filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Err: ", err)
			return err
		}

		if destDir == path {
			return nil
		}

		if !strings.Contains(path, currentPath) {
			translateFiles(filesToProcess, destDir)

			filesToProcess = []string{}
		}

		if info.IsDir() {
			currentPath = path
			return nil
		}

		filesToProcess = append(filesToProcess, path)

		return nil
	})

	if err != nil {
		fmt.Println("Err: ", err)
	}

	return err
}

func translateFiles(files []string, destDir string) {
	fmt.Println()
	fmt.Printf("directory output '%s'", destDir)
	fmt.Println()

	filesMap := make(map[string]*os.File)

	for _, originFilepath := range files {
		file, err := os.Open(originFilepath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		filename := removeExtension(getFilename(file.Name()))
		filesMap[filename] = file
	}

	parseActions(filesMap)

	fmt.Println()
	fmt.Println("------------------------")
	fmt.Println("------------------------")

	PrintMemUsage()
}

func removeExtension(path string) string {
	return strings.Split(path, ".")[0]
}

func getFilename(path string) string {
	return strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
}

func parseActions(filesMap map[string]*os.File) {
	file, ok := filesMap["actions"]
	if !ok {
		return
	}

	fmt.Println("parsing: ", file.Name())
	scanner := bufio.NewScanner(file)

	functionPattern := regexp.MustCompile(`\b(\w+)\((.+),\s(.*)\)\s{$`)
	commitNDispatchPattern := regexp.MustCompile(`\b(commit|dispatch)\(["|'](.+?)["|'],?\s?(.*)\)`)

	var lines []string
	var commitFuncs []string
	var dispatchFuncs []string

	for scanner.Scan() {
		line := scanner.Text()

		if match := functionPattern.FindStringSubmatch(line); match != nil {
			line = functionPattern.ReplaceAllString(line, "$1($3)")
		}

		if match := commitNDispatchPattern.FindStringSubmatch(line); match != nil {
			commitFuncs = append(commitFuncs, match[1])

			if strings.Contains(match[3], "root: true") {
				storeName := strings.Split(match[2], "/")[0]
				funcName := strings.Split(match[2], "/")[1]
				args := strings.Replace(match[3], ", { root: true }", "", 1)

				line = commitNDispatchPattern.ReplaceAllString(line, fmt.Sprintf("%s.%s(%s)", storeName, funcName, args))
			} else {
				line = commitNDispatchPattern.ReplaceAllString(line, "this.$2($3)")
			}

			line = commitNDispatchPattern.ReplaceAllString(line, "this.$2($3)")
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	err := os.WriteFile(file.Name(), []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("commit funcs: ", commitFuncs)
	fmt.Println("dispatch funcs: ", dispatchFuncs)
	fmt.Println()
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
