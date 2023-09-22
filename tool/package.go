package tool

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
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

	var mutationsLines = parseMutations(filesMap)
	var actionsLines = parseActions(filesMap)

	for index := len(actionsLines) - 1; index >= 0; index-- {
		// search latest line after close actions object
		if regexp.MustCompile(`^\};$`).FindStringSubmatch(actionsLines[index]) != nil {
			// insert mutations functions inside actions object
			for lineIndex, line := range mutationsLines {
				actionsLines = insert(actionsLines, index+lineIndex, line)
			}
		}
	}

	// get actions file to write lines
	file, ok := filesMap["actions"]
	if !ok {
		log.Fatal("actions file not found")
	}

	// write actions into output file
	err := os.WriteFile(file.Name(), []byte(strings.Join(actionsLines, "\n")), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func removeExtension(path string) string {
	return strings.Split(path, ".")[0]
}

func getFilename(path string) string {
	return strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
}

func insert(array []string, index int, value string) []string {
	if len(array) == index {
		return append(array, value)
	}
	array = append(array[:index+1], array[index:]...)
	array[index] = value
	return array
}

func parseMutations(filesMap map[string]*os.File) []string {
	file, ok := filesMap["mutations"]
	if !ok {
		return []string{}
	}

	fmt.Printf("parsing: %s\n\n", file.Name())
	scanner := bufio.NewScanner(file)

	var lines []string
	var insideMutations = false
	var index = -1
	var isFn = false

	mutObjPattern := regexp.MustCompile(`mutations\s=\s\{$|export\sdefault\s\{$`)
	fnPattern := regexp.MustCompile(`\b(\w+)\((\{[\w\s\,]+\}|\w+)((,\s*(.*))\)|\))((\:\s.+)?\s{)$`)
	endOfFunctionPattern := regexp.MustCompile(`\s\s\},?`)
	statePattern := regexp.MustCompile(`(state\.)(\w+)`)

	for scanner.Scan() {
		line := scanner.Text()

		if mutObjPattern.FindStringSubmatch(line) != nil {
			insideMutations = true
		}

		if insideMutations {
			if fnPattern.FindStringSubmatch(line) != nil {
				index++
				isFn = true
				line = fnPattern.ReplaceAllString(line, "$1($5)$6")
			}

			if statePattern.FindStringSubmatch(line) != nil {
				line = statePattern.ReplaceAllString(line, "this.$2")
			}
		}

		if isFn && index >= 0 {
			if len(lines) > index {
				lines[index] += "\n" + line
			} else {
				lines = append(lines, line)
			}
		}

		if endOfFunctionPattern.FindStringSubmatch(line) != nil {
			isFn = false
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines
}

func parseActions(filesMap map[string]*os.File) []string {
	file, ok := filesMap["actions"]
	if !ok {
		return []string{}
	}

	fmt.Printf("parsing: %s\n\n", file.Name())
	scanner := bufio.NewScanner(file)

	var lines []string
	var multiLineFn = []string{}
	// var multiLine = []string{}
	var importedStores = []string{}

	var commitStats []string
	var dispatchStats []string
	var actionsStats []string

	var actionCount int

	fnPattern := regexp.MustCompile(`\b(\w+)\((\{[\w\s\,]+\}|\w+)((,\s*(.*))\)|\))((\:\s.+)?\s{)$`)
	commitNDispatchPattern := regexp.MustCompile(`\b(commit|dispatch)\(["|'](.+?)["|'],?\s?(.*)\)`)
	statePattern := regexp.MustCompile(`(state\.)(\w+)`)

	for scanner.Scan() {
		line := scanner.Text()

		if regexp.MustCompile(`^\s{2}(async\s)?(\w)+\($`).FindStringSubmatch(line) != nil {
			multiLineFn = append(multiLineFn, strings.TrimSpace(line))

			continue
		} else if len(multiLineFn) > 0 {
			multiLineFn = append(multiLineFn, strings.TrimSpace(line))

			if regexp.MustCompile(`\s{2}\)\s{$`).FindStringSubmatch(line) != nil {
				line = fmt.Sprintf("  %s", strings.Join(multiLineFn, ""))

				multiLineFn = []string{}
			} else {
				continue
			}
		}

		if match := fnPattern.FindStringSubmatch(line); match != nil {
			actionCount++

			line = fnPattern.ReplaceAllString(line, "$1($5)$6")
			actionsStats = append(actionsStats, match[1])
		}

		if statePattern.FindStringSubmatch(line) != nil {
			line = statePattern.ReplaceAllString(line, "this.$2")
		}

		if match := commitNDispatchPattern.FindStringSubmatch(line); match != nil {
			switch match[1] {
			case "commit":
				commitStats = append(commitStats, match[2])
			case "dispatch":
				dispatchStats = append(dispatchStats, match[2])
			}

			if strings.Contains(match[3], "root: true") {
				fn := strings.Split(match[2], "/")
				storeFn := fmt.Sprintf("use%sStore", capitalizeByteSlice(fn[0]))
				storeName := fmt.Sprint(fn[0], "Store")
				fnName := fn[1]
				args := strings.Replace(match[3], ", { root: true }", "", 1)

				line = commitNDispatchPattern.ReplaceAllString(line, fmt.Sprintf("%s.%s(%s)", storeName, fnName, args))

				// get instance of root store
				line = fmt.Sprintf("\t\tconst %s = %s()\n%s", storeName, storeFn, line)

				if !slices.Contains(importedStores, storeName) {
					// create import statement of store
					importLine := fmt.Sprintf("import %s from '%s'", storeFn, fmt.Sprint("~/store/", fn[0], ".ts"))

					// TODO check that import statement not exists (sg: can use map to check)
					// add import statement to first line
					lines = append([]string{importLine}, lines...)
					importedStores = append(importedStores, storeName)
				}

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

	fmt.Println("action count: ", actionCount)
	if false {
		fmt.Println("actions: ", actionsStats)
		fmt.Println("commit functions: ", commitStats)
		fmt.Println("dispatch functions: ", dispatchStats)
	}

	return lines
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
