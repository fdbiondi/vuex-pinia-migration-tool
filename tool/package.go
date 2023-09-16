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
	checkMem := false

	if checkMem {
		PrintMemUsage()
	}

	err := filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Err: ", err)
			return err
		}

		if destDir == path {
			return nil
		}

		if !strings.Contains(path, currentPath) {
			fmt.Println("------------------------")
			translateFiles(filesToProcess)

			fmt.Println("------------------------")
			fmt.Println()

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

	if checkMem {
		PrintMemUsage()
	}

	return err
}

func translateFiles(files []string) {
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

	fmt.Printf("parsing: %s\n\n", file.Name())
	scanner := bufio.NewScanner(file)

	functionPattern := regexp.MustCompile(`\b(\w+)\((.+),\s(.*)\)(\s{)$`)
	commitNDispatchPattern := regexp.MustCompile(`\b(commit|dispatch)\(["|'](.+?)["|'],?\s?(.*)\)`)
	statePattern := regexp.MustCompile(`(state\.)(\w+)`)

	var lines []string
	var commitFns []string
	var dispatchFns []string
	var actions []string

	var startNewFn bool
	var actionCount int

	for scanner.Scan() {
		line := scanner.Text()

		if match := functionPattern.FindStringSubmatch(line); match != nil {
			startNewFn = true

			line = functionPattern.ReplaceAllString(line, "$1($3)$4")
			actions = append(actions, match[1])

		} else {
			startNewFn = false
		}

		if match := statePattern.FindStringSubmatch(line); match != nil {
			line = statePattern.ReplaceAllString(line, "this.$2")
		}

		if match := commitNDispatchPattern.FindStringSubmatch(line); match != nil {
			switch match[1] {
			case "commit":
				commitFns = append(commitFns, match[2])
			case "dispatch":
				dispatchFns = append(dispatchFns, match[2])
			}

			if strings.Contains(match[3], "root: true") {
				fn := strings.Split(match[2], "/")
				storeFn := fmt.Sprintf("%s%sStore", "use", capitalizeByteSlice(fn[0]))
				storeName := fmt.Sprint(fn[0], "Store")
				fnName := fn[1]
				args := strings.Replace(match[3], ", { root: true }", "", 1)

				line = commitNDispatchPattern.ReplaceAllString(line, fmt.Sprintf("%s.%s(%s)", storeName, fnName, args))

				// get instance of root store
				line = fmt.Sprintf("\t\tconst %s = %s()\n%s", storeName, storeFn, line)

				// create import statement of store
				importLine := fmt.Sprintf("import %s from '%s'\n", storeFn, fmt.Sprint("~/store/", fn[0], ".ts"))

				// TODO check that import statement not exists (sg: can use map to check)
				// add import statement to first line
				lines = append([]string{importLine}, lines...)
			} else {
				line = commitNDispatchPattern.ReplaceAllString(line, "this.$2($3)")
			}

			line = commitNDispatchPattern.ReplaceAllString(line, "this.$2($3)")
		}

		if startNewFn {
			actionCount++
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

	fmt.Println("action count: ", actionCount)
	fmt.Println("actions: ", actions)
	fmt.Println("commit functions: ", commitFns)
	fmt.Println("dispatch functions: ", dispatchFns)
}

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
