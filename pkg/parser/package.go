package parser

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var checkMem = false

func Execute(destDir string) error {
	filesInModule := []string{}
	currentPath := ""

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
			// pass current files inside a module to translation function
			fmt.Println("------------------------")
			translate(filesInModule)

			fmt.Println("------------------------")
			fmt.Println()

			// clean current module files
			filesInModule = []string{}
		}

		if info.IsDir() {
			currentPath = path
			return nil
		}

		// add files to current module
		filesInModule = append(filesInModule, path)

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

const CLOSE_FUNCTION_CURLY_BRACE = "  },"

const CLOSE_ACTIONS_PATTERN = `^\};$`

func translate(files []string) {
	filesMap := make(map[string]*os.File) // will have actions, mutations, state, getters keys

	// open and save files to the map
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
	var gettersLines = parseGetters(filesMap)
	var addedComma = false

	// append mutations into actions file
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

	// get getters file to write lines
	file, ok = filesMap["getters"]
	if !ok {
		log.Fatal("getters file not found")
	}

	// write getters into output file
	err = os.WriteFile(file.Name(), []byte(strings.Join(gettersLines, "\n")), 0644)
	if err != nil {
		log.Fatal(err)
	}

	// remove mutations file
	err = os.Remove(filesMap["mutations"].Name())
	if err != nil {
		log.Fatal(err)
	}

	// create module entrypoint
	var storeFilename = strings.Replace(filesMap["actions"].Name(), "actions", "index", 1)

	pathSlice := strings.Split(storeFilename, "/")
	storeName := pathSlice[len(pathSlice)-2]

	err = createEntryPoint("./templates/store-index.tmpl", storeFilename, map[string]string{
		"storeName":          storeName,
		"storeNameTitleCase": cases.Title(language.English).String(storeName),
	})
	if err != nil {
		log.Fatal(err)
	}

}

func createEntryPoint(templatePath string, outputPath string, values map[string]string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	tmp, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmp.Execute(&buf, values); err != nil {
		return err
	}

	err = os.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		return errors.New("error wrinting file -> " + err.Error())
	}

	return nil
}
