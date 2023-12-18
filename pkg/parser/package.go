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
		m.path = path
		return nil
	} else if m.path == "" {
		m.path = regexp.MustCompile(`(.*)\/(.*)$`).ReplaceAllString(path, "$1")
	}

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
			m.translate()

			if m.parentName == "" {
				fmt.Printf("Created %s store\n", modName)
			} else {
				fmt.Printf("Created %s store\n", strings.Join([]string{m.parentName, modName}, "/"))
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

const CLOSE_FUNCTION_CURLY_BRACE = "  },"

const CLOSE_ACTIONS_PATTERN = `^\};$`

const IMPORT_PATTERN = `^import.*$`

var replaceImportPattern = regexp.MustCompile(`^import\s((\w+)|\{(.*)\})\sfrom\s(('|").*('|"))(;|)$`)

func (m *Module) translate() {
	filesMap := make(map[string]*os.File) // will have actions, mutations, state, getters keys

	// open and save files to the map
	for _, originFilepath := range m.files {
		file, err := os.Open(originFilepath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		filename := removeExtension(getFilename(file.Name()))
		filesMap[filename] = file
	}

	var mutationsLines, mutationsImportLines = parseMutations(filesMap)
	var actionsLines = parseActions(filesMap)
	var gettersLines = parseGetters(filesMap)
	var addedComma = false
	var migrated = false

	// append mutations into actions file
	if len(mutationsLines) > 0 {
		for index := len(actionsLines) - 1; index >= 0; index-- {
			// search latest line after close actions object
			if regexp.MustCompile(CLOSE_ACTIONS_PATTERN).FindStringSubmatch(actionsLines[index]) != nil {

				// this fixes the last action adding a comma at the end of it
				if !addedComma {
					actionsLines[index-1] = CLOSE_FUNCTION_CURLY_BRACE

					addedComma = true
				}

				// insert mutations functions inside actions object
				for lineIndex, line := range mutationsLines {
					actionsLines = insertLine(actionsLines, index+(lineIndex*2+0), "")
					actionsLines = insertLine(actionsLines, index+(lineIndex*2+1), line)
				}

				break
			}
		}
	}

	var replacedImports = []int{}
	var lastImportIndex = 0

	if len(mutationsImportLines) > 0 {
		for index, actionLine := range actionsLines {

			if actMatch := replaceImportPattern.FindStringSubmatch(actionLine); actMatch != nil {

				for mutIndex, mutLine := range mutationsImportLines {

					if slices.Contains(replacedImports, mutIndex) {
						continue
					}

					if mutMatch := replaceImportPattern.FindStringSubmatch(mutLine); mutMatch != nil {

						// check same file imported and is not a default import
						if mutMatch[4] == actMatch[4] && mutMatch[2] == "" {
							// remove mutations related import
							importValue := regexp.MustCompile(`(\w+)?(Mutation|mutation)(\w+)?(,|)`).ReplaceAllString(mutMatch[3], "")

							line := replaceImportPattern.ReplaceAllString(actionLine, fmt.Sprintf("import {$3,%s} from $4$7", importValue))

							actionsLines[index] = line
							replacedImports = append(replacedImports, mutIndex)
						}
					}
				}

				lastImportIndex = index
			}

			if len(replacedImports) == len(mutationsImportLines) {
				break
			}
		}

		if len(replacedImports) != len(mutationsImportLines) {
			for lineIndex, line := range mutationsImportLines {
				if slices.Contains(replacedImports, lineIndex) {
					continue
				}

				actionsLines = insertLine(actionsLines, lastImportIndex+lineIndex+1, line)
			}
		}
	}

	// get actions file to write lines
	file, ok := filesMap["actions"]
	if ok {
		// write actions into output file
		err := os.WriteFile(file.Name(), []byte(strings.Join(actionsLines, "\n")), 0644)
		if err != nil {
			log.Fatal(err)
		}

		migrated = true
	}

	// get getters file to write lines
	file, ok = filesMap["getters"]
	if ok {
		// write getters into output file
		err := os.WriteFile(file.Name(), []byte(strings.Join(gettersLines, "\n")), 0644)
		if err != nil {
			log.Fatal(err)
		}

		migrated = true
	}

	// remove mutations file
	file, ok = filesMap["mutations"]
	if ok {
		err := os.Remove(file.Name())
		if err != nil {
			log.Fatal(err)
		}

		migrated = true
	}

	if !migrated {
		return
	}

	// create module entrypoint
	_, err := createIndex(filesMap)
	if err != nil {
		log.Fatal(err)
	}
}
