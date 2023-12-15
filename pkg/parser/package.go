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

func Execute(destDir string) error {
	filesInModule := []string{}
	filesInDir := []string{}
	filesInSubModules := map[string][]string{}
	subModules := []string{}
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
		if len(append(filesInDir, subModules...)) == 0 {
			entries, _ := os.ReadDir(currentPath)
			for _, e := range entries {

				fileInfo, err := os.Stat(strings.Join([]string{currentPath, e.Name()}, "/"))
				if err != nil {
					// TODO skips files with errors for now
					continue
				}

				if fileInfo.IsDir() {
					subModules = append(subModules, e.Name())
				} else {
					filesInDir = append(filesInDir, e.Name())
				}
			}
		}

		currModule := regexp.MustCompile(`(.*)\/(.*)\/(.*)$`).ReplaceAllString(path, "$2")
		// modulePath := regexp.MustCompile(`(.*)\/(.*)$`).ReplaceAllString(path, "$1")

		if slices.Contains(subModules, currModule) {
			filesInSubModules[currModule] = append(filesInSubModules[currModule], path)

			return nil
		}

		// add files to current module
		filesInModule = append(filesInModule, path)

		// checking last file inside the directory
		if filesInDir[len(filesInDir)-1] == info.Name() {

			printOutput(path, func() {
				translate(filesInModule)
			})

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

func printOutput(path string, fn func()) {
	tag := fmt.Sprintf("--------------------%s--------------------", strings.Split(path, "/")[len(strings.Split(path, "/"))-2])

	if Verbose {
		fmt.Println(tag)
	}

	fn()

	if Verbose {
		for range tag {
			fmt.Printf("-")
		}

		fmt.Println()
		fmt.Println()
	}
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
	var migrated = false

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
	storeName, err := createEntryPoint(filesMap)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created %s store\n", storeName)
}
