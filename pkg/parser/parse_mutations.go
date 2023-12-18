package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
)

var mutPattern = map[string]*regexp.Regexp{
	string("object"):       regexp.MustCompile(`(mutations(:\s\w+)?\s=\s\{)|(export\sdefault\s\{)`),
	string("function"):     regexp.MustCompile(`\b(\w+)\((\{[\w\s\,]+\}|\w+)((,\s*(.*))\)|\))((\:\s.+)?\s{)$`),
	string("function_end"): regexp.MustCompile(`(?m)^\s\s\},?$`),
	string("state_prop_1"): regexp.MustCompile(`(state\.)(\w+)`),
	string("state_prop_2"): regexp.MustCompile(`(state(,|))`),
	string("import"):       regexp.MustCompile(`^import.*$`),
	string("import_store"): regexp.MustCompile(`~/store/`),
}

func parseMutations(filesMap map[string]*os.File) ([]string, []string) {
	file, ok := filesMap["mutations"]
	if !ok {
		return []string{}, []string{}
	}

	if Verbose {
		fmt.Printf("parsing: %s\n", file.Name())
	}
	scanner := bufio.NewScanner(file)

	var lines []string
	var importLines []string
	var insideMutations = false
	var index = -1
	var isFn = false

	for scanner.Scan() {
		line := scanner.Text()

		if mutPattern["import"].FindStringSubmatch(line) != nil {
			if mutPattern["import_store"].FindStringSubmatch(line) != nil {
				line = mutPattern["import_store"].ReplaceAllString(line, "~/stores/")
			}

			importLines = append(importLines, line)
		}

		if mutPattern["object"].FindStringSubmatch(line) != nil {
			insideMutations = true
		}

		if insideMutations {
			if mutPattern["function"].FindStringSubmatch(line) != nil {
				index++
				isFn = true
				line = mutPattern["function"].ReplaceAllString(line, "$1($5)$6")
			}

			if mutPattern["state_prop_1"].FindStringSubmatch(line) != nil {
				line = mutPattern["state_prop_1"].ReplaceAllString(line, "this.$2")
			} else if mutPattern["state_prop_2"].FindStringSubmatch(line) != nil {
				line = mutPattern["state_prop_2"].ReplaceAllString(line, "this$2")
			}
		}

		if isFn && index >= 0 {
			if len(lines) > index {
				lines[index] += "\n" + line
			} else {
				lines = append(lines, line)
			}
		}

		if mutPattern["function_end"].FindStringSubmatch(line) != nil {
			isFn = false
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines, importLines
}
