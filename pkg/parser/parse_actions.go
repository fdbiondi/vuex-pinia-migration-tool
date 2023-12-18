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
	string("function"):                  regexp.MustCompile(`\b(\w+)\((\{[\w\s\,]+\}|\w+)((,\s*(.*))\)|\))((\:\s.+)?\s{)$`),
	string("commit_dispatch"):           regexp.MustCompile(`\b(commit|dispatch)\(["|'](.+?)["|'],?\s?(.*)\)`),
	string("state_prop"):                regexp.MustCompile(`(state\.)(\w+)`),
	string("commit_dispatch_lines"):     regexp.MustCompile(`\b(dispatch|commit)\('(.*)',\s(\{)$|\b(dispatch|commit)\($`),
	string("commit_dispatch_lines_end"): regexp.MustCompile(`.*\);$`),
	string("function_lines"):            regexp.MustCompile(`^\s{2}(async\s)?(\w)+\($`),
	string("function_lines_end"):        regexp.MustCompile(`\s{2}\)\s{$`),
	string("getter_call"):               regexp.MustCompile(`(.*)getters(\.\w+.*)`),
}

func parseActions(filesMap map[string]*os.File) []string {
	file, ok := filesMap["actions"]
	if !ok {
		return []string{}
	}

	if Verbose {
		fmt.Printf("parsing: %s\n", file.Name())
	}
	scanner := bufio.NewScanner(file)

	var lines []string
	var multiLineAction = []string{}
	var multiLineFnCall = []string{}
	var importedStores = []string{}
	var intantiatedStores = []string{}

	for scanner.Scan() {
		line := scanner.Text()

		if actionPattern["function_lines"].FindStringSubmatch(line) != nil {
			multiLineAction = append(multiLineAction, strings.TrimSpace(line))

			continue
		} else if len(multiLineAction) > 0 {
			multiLineAction = append(multiLineAction, strings.TrimSpace(line))

			if actionPattern["function_lines_end"].FindStringSubmatch(line) != nil {
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

		if match := getterPattern["getter_call"].FindStringSubmatch(line); match != nil {
			line = getterPattern["getter_call"].ReplaceAllString(line, "$1this$2")
		}

		if actionPattern["commit_dispatch_lines"].FindStringSubmatch(line) != nil && len(multiLineFnCall) == 0 {
			multiLineFnCall = append(multiLineFnCall, strings.TrimSpace(line))

			continue
		} else if len(multiLineFnCall) > 0 {
			multiLineFnCall = append(multiLineFnCall, strings.TrimSpace(line))

			if actionPattern["commit_dispatch_lines_end"].FindStringSubmatch(line) != nil {
				line = fmt.Sprintf("      %s", strings.Join(multiLineFnCall, ""))

				multiLineFnCall = []string{}
			} else {
				continue
			}
		}

		if match := actionPattern["commit_dispatch"].FindStringSubmatch(line); match != nil {

			if strings.Contains(match[3], "root: true") {
				// should import another store
				fn := strings.Split(match[2], "/")
				storeFn := fmt.Sprintf("use%sStore", capitalizeByteSlice(fn[0]))
				storeName := fmt.Sprint(fn[0], "Store")
				fnName := fn[1]
				args := strings.Replace(match[3], ", { root: true }", "", 1)
				args = strings.Replace(args, ",{ root: true }", "", 1)

				line = actionPattern["commit_dispatch"].ReplaceAllString(line, fmt.Sprintf("%s.%s(%s)", storeName, fnName, args))

				if !slices.Contains(intantiatedStores, storeName) {
					// get instance of root store
					line = fmt.Sprintf("\t\tconst %s = %s()\n%s", storeName, storeFn, line)
					intantiatedStores = append(intantiatedStores, storeName)
				}

				if !slices.Contains(importedStores, storeName) {
					// create import statement of store
					importLine := fmt.Sprintf("import { %s } from '%s'", storeFn, fmt.Sprint("~/stores/", fn[0]))

					// add import statement to first line
					lines = append([]string{importLine}, lines...)
					importedStores = append(importedStores, storeName)
				}

			} else {
				line = actionPattern["commit_dispatch"].ReplaceAllString(line, "this.$2($3)")
			}

			line = actionPattern["commit_dispatch"].ReplaceAllString(line, "this.$2($3)")
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines
}
