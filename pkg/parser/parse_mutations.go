package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
)

var mutPattern = map[string]*regexp.Regexp{
	string("object"):       regexp.MustCompile(`mutations\s=\s\{$|export\sdefault\s\{$`),
	string("function"):     regexp.MustCompile(`\b(\w+)\((\{[\w\s\,]+\}|\w+)((,\s*(.*))\)|\))((\:\s.+)?\s{)$`),
	string("function_end"): regexp.MustCompile(`\s\s\},?`),
	string("state_prop"):   regexp.MustCompile(`(state\.)(\w+)`),
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

	for scanner.Scan() {
		line := scanner.Text()

		if mutPattern["object"].FindStringSubmatch(line) != nil {
			insideMutations = true
		}

		if insideMutations {
			if mutPattern["function"].FindStringSubmatch(line) != nil {
				index++
				isFn = true
				line = mutPattern["function"].ReplaceAllString(line, "$1($5)$6")
			}

			if mutPattern["state_prop"].FindStringSubmatch(line) != nil {
				line = mutPattern["state_prop"].ReplaceAllString(line, "this.$2")
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

	return lines
}
