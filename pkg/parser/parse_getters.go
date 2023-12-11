package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
)

var getterPattern = map[string]*regexp.Regexp{
	string("function"):           regexp.MustCompile(`\b(\w+).*\((state)(,\s.*)\)`),
	string("function_lines_end"): regexp.MustCompile(`\s{2}\},$`),
	string("getter_call"):        regexp.MustCompile(`getters\.(\w*)`),
	string("root_state_call"):    regexp.MustCompile(`rootState\.(\w*)\.(\w*)`),
	string("root_getter_call"):   regexp.MustCompile(`rootGetters\[("|')((\w+)\/(\w+))("|')\]`),
}

func parseGetters(filesMap map[string]*os.File) []string {
	file, ok := filesMap["getters"]
	if !ok {
		return []string{}
	}

	fmt.Printf("parsing: %s\n\n", file.Name())
	scanner := bufio.NewScanner(file)

	var lines []string
	var functionStarted = false
	var functionLines = []string{}
	var importedStores = []string{}
	var intantiatedStores = []string{}

	for scanner.Scan() {
		line := scanner.Text()

		if match := getterPattern["function"].FindStringSubmatch(line); match != nil {
			line = getterPattern["function"].ReplaceAllString(line, "$1($2)")

			// allows to check if a store has been declared
			intantiatedStores = []string{}
			// allows to add lines at the beginning of the function
			functionStarted = true

			lines = append(lines, line)
			continue
		}

		if match := getterPattern["getter_call"].FindStringSubmatch(line); match != nil {
			line = getterPattern["getter_call"].ReplaceAllString(line, "this.$1")
		}

		if match := getterPattern["root_state_call"].FindStringSubmatch(line); match != nil {
			// get store name and function name to create instance
			storeName := fmt.Sprintf("%s%sStore", match[1], capitalizeByteSlice(match[2]))
			storeFn := fmt.Sprintf("use%s", capitalizeByteSlice(storeName))

			// replace original line
			var lineReplace = fmt.Sprintf("%s.$$state", storeName)

			line = getterPattern["root_state_call"].ReplaceAllString(line, lineReplace)

			if !slices.Contains(intantiatedStores, storeName) {
				// get instance of root store
				defLine := fmt.Sprintf("\t\tconst %s = %s()", storeName, storeFn)

				functionLines = append([]string{defLine}, functionLines...)

				intantiatedStores = append(intantiatedStores, storeName)
			}

			if !slices.Contains(importedStores, storeName) {
				// create import statement of store
				storeFilename := fmt.Sprintf("%s-%s", match[1], match[2])
				importLine := fmt.Sprintf("import %s from '%s'", storeFn, fmt.Sprint("~/store/", storeFilename, ".ts"))

				// add import statement to first line
				lines = append([]string{importLine}, lines...)
				importedStores = append(importedStores, storeName)
			}
		}

		if match := getterPattern["root_getter_call"].FindStringSubmatch(line); match != nil {

			storeFn := fmt.Sprintf("use%sStore", capitalizeByteSlice(match[3]))
			storeName := fmt.Sprint(match[3], "Store")
			fnName := match[4]
			// args := strings.Replace(match[3], ", { root: true }", "", 1)

			line = getterPattern["root_getter_call"].ReplaceAllString(line, fmt.Sprintf("%s.%s", storeName, fnName))

			if !slices.Contains(intantiatedStores, storeName) {
				// get instance of root store
				defLine := fmt.Sprintf("\t\tconst %s = %s()", storeName, storeFn)

				functionLines = append([]string{defLine}, functionLines...)

				intantiatedStores = append(intantiatedStores, storeName)
			}

			if !slices.Contains(importedStores, storeName) {
				// create import statement of store
				importLine := fmt.Sprintf("import %s from '%s'", storeFn, fmt.Sprint("~/store/", storeName, ".ts"))

				// add import statement to first line
				lines = append([]string{importLine}, lines...)
				importedStores = append(importedStores, storeName)
			}
		}

		if match := getterPattern["function_lines_end"].FindStringSubmatch(line); match != nil {
			functionStarted = false
		}

		if functionStarted {
			functionLines = append(functionLines, line)
			// if adding function content go to next line
			continue
		}

		if len(functionLines) > 0 {
			lines = append(lines, functionLines...)
			functionLines = []string{}
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines
}
