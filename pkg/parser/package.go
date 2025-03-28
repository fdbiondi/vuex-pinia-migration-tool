package parser

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var (
	Debug   = false
	Verbose = false
)

type Module struct {
	files      []string
	dirList    []string
	subModules map[string]string
	path       string
	outputDir  string
	parentName string
}

var (
	subModules = []string{}
)

func NewModule(outputDir string) Module {
	return Module{
		outputDir:  outputDir,
		parentName: "",
		subModules: make(map[string]string),
		path:       "",
	}
}

func NewSubModule(outputDir string, parentName string) Module {
	return Module{
		outputDir:  outputDir,
		parentName: parentName,
		subModules: make(map[string]string),
		path:       "",
	}
}

func (m *Module) Parse() error {
	if Debug {
		PrintMemUsage()
	}

	err := filepath.Walk(m.outputDir, m.walk)

	if err != nil && Verbose {
		fmt.Println("Err: ", err)
	}

	if Debug {
		PrintMemUsage()
	}

	return err
}

func (m *Module) walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		if Verbose {
			fmt.Println("Err: ", err)
		}
		return err
	}

	if m.outputDir == path {
		return nil
	}

	// TODO: only works for directories for now
	if info.IsDir() {
		return nil
	}

	m.path = regexp.MustCompile(`(.*)\/(.*)$`).ReplaceAllString(path, "$1")

	// save all file names to check later the last file in dir
	if len(append(m.dirList, subModules...)) == 0 {
		entries, _ := os.ReadDir(m.path)
		for _, e := range entries {

			fileInfo, err := os.Stat(strings.Join([]string{m.path, e.Name()}, "/"))
			if err != nil {
				// TODO skips files with errors for now
				continue
			}

			if fileInfo.IsDir() {
				subModules = append(subModules, e.Name())
			} else {
				m.dirList = append(m.dirList, e.Name())
			}
		}
	}

	modName := regexp.MustCompile(`(.*)\/(.*)\/(.*)$`).ReplaceAllString(path, "$2")

	if slices.Contains(subModules, modName) {
		_, ok := m.subModules[modName]
		if !ok {
			m.subModules[modName] = regexp.MustCompile(`(.*)\/(.*)$`).ReplaceAllString(path, "$1")
		}

		return nil
	}

	// add files to current module
	m.files = append(m.files, path)

	// checking last file inside the directory
	if m.dirList[len(m.dirList)-1] == info.Name() {

		printOutput(path, func() {
			migrated := m.translate()

			if migrated {
				if m.parentName == "" {
					fmt.Printf("Created %s store\n", modName)
				} else {
					fmt.Printf("Created %s store\n", strings.Join([]string{m.parentName, modName}, "/"))
				}
			}
		})

		// clean current module files
		m.files = []string{}
		// clean files in dir
		m.dirList = []string{}
		// clean sub modules
		subModules = []string{}

		// when finishes current module, parse sub modules
		for _, subModPath := range m.subModules {
			var parentName string

			if m.parentName == "" {
				parentName = strings.Join([]string{modName}, "/")
			} else {
				parentName = strings.Join([]string{m.parentName, modName}, "/")
			}

			var subModule = NewSubModule(subModPath, parentName)

			subModule.Parse()
		}

		m.subModules = make(map[string]string)
	}

	return nil
}

func (m *Module) translate() bool {
	filesMap := make(map[string]*os.File) // will have actions, mutations, state, getters keys

	// open and save files to the map
	for _, originFilepath := range m.files {
		file, err := os.Open(originFilepath)
		if err != nil {
			continue
		}
		defer file.Close()

		filename := removeExtension(getFilename(file.Name()))
		filesMap[filename] = file
	}

	actionsPath, _ := checkActionsFile(filesMap)
	if actionsPath != "" {
		file, _ := os.Open(actionsPath)
		filesMap["actions"] = file

		defer file.Close()
	}

	var mutationsLines, mutationsImportLines = parseMutations(filesMap)
	var actionsLines = parseActions(filesMap)
	var gettersLines = parseGetters(filesMap)
	var migrated = []string{}

	appendLinesToObj(&actionsLines, &mutationsLines)
	appendImports(&actionsLines, &mutationsImportLines)

	// get actions file to write lines
	file, ok := filesMap["actions"]
	if ok {
		// write actions into output file
		err := os.WriteFile(file.Name(), []byte(strings.Join(actionsLines, "\n")), 0644)
		if err != nil {
			log.Fatal(err)
		}

		migrated = append(migrated, "actions")
	}

	// get getters file to write lines
	file, ok = filesMap["getters"]
	if ok {
		// write getters into output file
		err := os.WriteFile(file.Name(), []byte(strings.Join(gettersLines, "\n")), 0644)
		if err != nil {
			log.Fatal(err)
		}

		migrated = append(migrated, "getters")
	}

	file, ok = filesMap["mutations"]
	if ok {
		// remove mutations file
		os.Remove(file.Name())

		for _, ext := range []string{".ts", ".js"} {
			os.Remove(strings.Replace(file.Name(), ext, ".spec.ts", 1))
			os.Remove(strings.Replace(file.Name(), ext, ".spec.js", 1))
		}

		migrated = append(migrated, "mutations")
	}

	if len(migrated) == 0 {
		return false
	}

	// set template type for index file
	var templateType = DEFAULT_TEMPLATE
	if !slices.Contains(migrated, "actions") && !slices.Contains(migrated, "getters") && !slices.Contains(migrated, "mutations") {
		templateType = STATE_ONLY_TEMPLATE
	} else if !slices.Contains(migrated, "actions") && !slices.Contains(migrated, "mutations") {
		templateType = NO_ACTIONS_TEMPLATE
	} else if !slices.Contains(migrated, "getters") {
		templateType = NO_GETTERS_TEMPLATE
	}

	// create module entrypoint
	var templatePath = getTemplatePath(filesMap, "index")
	var storeName = strings.Split(templatePath, "/")[len(strings.Split(templatePath, "/"))-2]
	var values = map[string]string{
		"storeName":          storeName,
		"storeNameTitleCase": kebabToCamelCase(storeName, true),
	}

	err := createTemplate(templateType, templatePath, values)

	return err == nil
}

func checkActionsFile(filesMap map[string]*os.File) (string, error) {
	_, actionsFileOk := filesMap["actions"]
	_, mutationsFileOk := filesMap["mutations"]
	if !actionsFileOk && mutationsFileOk {
		var templatePath = strings.Replace(getTemplatePath(filesMap, "index"), "index", "actions", 1)

		// create actions file
		err := createTemplate(ACTIONS_EMPTY_TEMPLATE, templatePath, map[string]string{})
		if err != nil {
			return "", err
		}

		return templatePath, nil
	}

	return "", nil
}

func appendLinesToObj(lines *[]string, linesToAppend *[]string) {
	const CLOSE_OBJ_LINE = `^\};$`
	const OBJ_START_LINE = `const\s\w+(:\s\w+)?\s=\s{$`
	const CLOSE_FUNCTION_CURLY_BRACE = "  },"

	if len(*linesToAppend) == 0 {
		return
	}

	// start from the last line
	for index := len(*lines) - 1; index >= 0; index-- {
		// search latest line after close the object
		if regexp.MustCompile(CLOSE_OBJ_LINE).FindStringSubmatch((*lines)[index]) != nil {

			if regexp.MustCompile(OBJ_START_LINE).FindStringSubmatch((*lines)[index-2]) == nil {
				// this fixes the last line adding a comma at the end of it
				// only in the case that not using the template file
				(*lines)[index-1] = CLOSE_FUNCTION_CURLY_BRACE
			}

			// insert lines inside object
			for lineIndex, line := range *linesToAppend {
				*lines = insertLine(*lines, index+(lineIndex*2+0), "")
				*lines = insertLine(*lines, index+(lineIndex*2+1), line)
			}

			break
		}
	}
}

func appendImports(lines *[]string, importLines *[]string) {
	var replaceImportPattern = regexp.MustCompile(`^import\s((\w+)|\{(.*)\})\sfrom\s(('|").*('|"))(;|)$`)
	var namedValuePattern = regexp.MustCompile(`(\w+)?(Mutation|mutation)(\w+)?(,|)`)

	var replacedImports = []int{}
	var lastImportIndex = 0

	if len(*importLines) == 0 || len(*lines) == 0 {
		return
	}

	for index, line := range *lines {

		if match := replaceImportPattern.FindStringSubmatch(line); match != nil {

			for importIndex, importLine := range *importLines {

				if slices.Contains(replacedImports, importIndex) {
					continue
				}

				if importMatch := replaceImportPattern.FindStringSubmatch(importLine); importMatch != nil {

					// check same file imported and is not a default import
					if importMatch[4] == match[4] && importMatch[2] == "" {
						// remove mutations related import
						importValue := namedValuePattern.ReplaceAllString(importMatch[3], "")
						line := replaceImportPattern.ReplaceAllString(line, fmt.Sprintf("import {$3,%s} from $4$7", importValue))

						(*lines)[index] = line
						replacedImports = append(replacedImports, importIndex)
					}
				}
			}

			lastImportIndex = index
		}

		if len(replacedImports) == len(*importLines) {
			break
		}
	}

	if len(replacedImports) != len(*importLines) {
		for lineIndex, line := range *importLines {
			if slices.Contains(replacedImports, lineIndex) {
				continue
			}

			if namedValuePattern.FindStringSubmatch(line) != nil {
				continue
			}

			*lines = insertLine(*lines, lastImportIndex+lineIndex, line)
		}
	}
}
