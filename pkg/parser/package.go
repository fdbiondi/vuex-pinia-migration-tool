package parser

import (
	"bytes"
	"embed"
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

//go:embed templates/index.tmpl
var indexTmpl embed.FS

var (
	Debug   = false
	Verbose = false
)

func Execute(destDir string) error {
	filesInModule := []string{}
	filesInDir := []string{}
	currentPath := ""

	if Debug {
		PrintMemUsage()
	}

	err := filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if Verbose {
				fmt.Println("Err: ", err)
			}
			return err
		}

		if destDir == path {
			return nil
		}

		// TODO: only works for directories for now
		if info.IsDir() {
			currentPath = path
			return nil
		}

		// save all file names to check later the last file in dir
		if len(filesInDir) == 0 {
			entries, _ := os.ReadDir(currentPath)
			for _, e := range entries {
				filesInDir = append(filesInDir, e.Name())
			}
		}

		// add files to current module
		filesInModule = append(filesInModule, path)

		// checking last file inside the directory
		if filesInDir[len(filesInDir)-1] == info.Name() {
			// pass current files inside a module to translation function
			tag := fmt.Sprintf("--------------------%s--------------------", strings.Split(path, "/")[len(strings.Split(path, "/"))-2])

			if Verbose {
				fmt.Println(tag)
			}
			translate(filesInModule)

			if Verbose {
				for range tag {
					fmt.Printf("-")
				}

				fmt.Println()
				fmt.Println()
			}

			// clean current module files
			filesInModule = []string{}
			// clean files in dir
			filesInDir = []string{}
		}

		return nil
	})

	if err != nil && Verbose {
		fmt.Println("Err: ", err)
	}

	if Debug {
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
	if ok {
		// write actions into output file
		err := os.WriteFile(file.Name(), []byte(strings.Join(actionsLines, "\n")), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	// get getters file to write lines
	file, ok = filesMap["getters"]
	if ok {
		// write getters into output file
		err := os.WriteFile(file.Name(), []byte(strings.Join(gettersLines, "\n")), 0644)
		if err != nil {
			log.Fatal(err)
		}

	}

	// remove mutations file
	err := os.Remove(filesMap["mutations"].Name())

	// create module entrypoint
	var storeFilename = strings.Replace(filesMap["actions"].Name(), "actions", "index", 1)

	pathSlice := strings.Split(storeFilename, "/")
	storeName := pathSlice[len(pathSlice)-2]

	data, err := indexTmpl.ReadFile("templates/index.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	var values = map[string]string{
		"storeName":          storeName,
		"storeNameTitleCase": cases.Title(language.English).String(storeName),
	}

	err = parseTemplate(string(data), storeFilename, values)
	if err != nil {
		log.Fatal(err)
	}
}

func parseTemplate(templateStr string, outputPath string, values map[string]string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl := template.Must(template.New("template").Parse(templateStr))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, values); err != nil {
		return err
	}

	err = os.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		return errors.New("error wrinting file -> " + err.Error())
	}

	return nil
}
