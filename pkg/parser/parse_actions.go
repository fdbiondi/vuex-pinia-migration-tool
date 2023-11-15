package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"
)

var actionPattern = map[string]*regexp.Regexp{
	string("function"):        regexp.MustCompile(`\b(\w+)\((\{[\w\s\,]+\}|\w+)((,\s*(.*))\)|\))((\:\s.+)?\s{)$`),
	string("commit_dispatch"): regexp.MustCompile(`\b(commit|dispatch)\(["|'](.+?)["|'],?\s?(.*)\)`),
	string("state_prop"):      regexp.MustCompile(`(state\.)(\w+)`),
	string("commit_lines"):    regexp.MustCompile(`\b(commit)\('(\w+)',\s(.*)`),
	string("dispatch_lines"):  regexp.MustCompile(`\b(dispatch)\('(\w+)',\s(.*)`),
}

var multiLineActionPattern = map[string]*regexp.Regexp{
	string("begin"): regexp.MustCompile(`^\s{2}(async\s)?(\w)+\($`),
	string("end"):   regexp.MustCompile(`\s{2}\)\s{$`),
}

func parseActions(filesMap map[string]*os.File) []string {
	file, ok := filesMap["actions"]
	if !ok {
		return []string{}
	}

	fmt.Printf("parsing: %s\n\n", file.Name())
	scanner := bufio.NewScanner(file)

	var lines []string
	var multiLineAction = []string{}
	var importedStores = []string{}
	var intantiatedStores = []string{}

	var actionsStats = []string{}

	for scanner.Scan() {
		line := scanner.Text()

		if multiLineActionPattern["begin"].FindStringSubmatch(line) != nil {
			multiLineAction = append(multiLineAction, strings.TrimSpace(line))

			continue
		} else if len(multiLineAction) > 0 {
			multiLineAction = append(multiLineAction, strings.TrimSpace(line))

			if multiLineActionPattern["end"].FindStringSubmatch(line) != nil {
				line = fmt.Sprintf("  %s", strings.Join(multiLineAction, ""))

				multiLineAction = []string{}
			} else {
				continue
			}
		}

		if match := actionPattern["function"].FindStringSubmatch(line); match != nil {

			line = actionPattern["function"].ReplaceAllString(line, "$1($5)$6")

			intantiatedStores = []string{}
		}

		if actionPattern["state_prop"].FindStringSubmatch(line) != nil {
			line = actionPattern["state_prop"].ReplaceAllString(line, "this.$2")
		}

		if match := actionPattern["commit_dispatch"].FindStringSubmatch(line); match != nil {

			if strings.Contains(match[3], "root: true") {
				// should import another store
				fn := strings.Split(match[2], "/")
				storeFn := fmt.Sprintf("use%sStore", capitalizeByteSlice(fn[0]))
				storeName := fmt.Sprint(fn[0], "Store")
				fnName := fn[1]
				args := strings.Replace(match[3], ", { root: true }", "", 1)

				line = actionPattern["commit_dispatch"].ReplaceAllString(line, fmt.Sprintf("%s.%s(%s)", storeName, fnName, args))

				if !slices.Contains(intantiatedStores, storeName) {
					// get instance of root store
					line = fmt.Sprintf("\t\tconst %s = %s()\n%s", storeName, storeFn, line)
					intantiatedStores = append(intantiatedStores, storeName)
				}

				if !slices.Contains(importedStores, storeName) {
					// create import statement of store
					importLine := fmt.Sprintf("import %s from '%s'", storeFn, fmt.Sprint("~/store/", fn[0], ".ts"))

					// TODO check that import statement not exists (sg: can use map to check)
					// add import statement to first line
					lines = append([]string{importLine}, lines...)
					importedStores = append(importedStores, storeName)
				}

			} else {
				line = actionPattern["commit_dispatch"].ReplaceAllString(line, "this.$2($3)")
			}

			line = actionPattern["commit_dispatch"].ReplaceAllString(line, "this.$2($3)")
		}

		if match := actionPattern["commit_lines"].FindStringSubmatch(line); match != nil {
			line = actionPattern["commit_lines"].ReplaceAllString(line, "this.$2($3")

			actionsStats = append(actionsStats, match[2])

			intantiatedStores = []string{}
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("actions: ", actionsStats)

	return lines
}
