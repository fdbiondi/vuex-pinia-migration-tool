package tool

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func Transform(dest_dir string) {
	current_path := ""
	files_to_process := []string{}

	err := filepath.Walk(dest_dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Err: ", err)
			return err
		}

		if dest_dir == path {
			return nil
		}

		if !strings.Contains(path, current_path) {
			translateFiles(files_to_process, dest_dir)

			files_to_process = []string{}
			fmt.Println("")

		}

		if info.IsDir() {
			// current_dir = strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
			current_path = path
			fmt.Println("")
			// fmt.Println("current_dir: ", strings.ToUpper(current_dir))

			return nil
		}

		files_to_process = append(files_to_process, path)

		return nil
	})

	if err != nil {
		fmt.Println("Err: ", err)
	}
}

func translateFiles(files []string, dest_dir string) {
	fmt.Printf("directory output '%s'\n", dest_dir)
	fmt.Println("")

	for _, origin_filepath := range files {
		fmt.Println("filename => ", origin_filepath)
		fmt.Println("------------------------")

		readFile(origin_filepath)

		fmt.Println("------------------------")
		fmt.Println("")
		// break
	}
}

func readFile(path string) {
	// files_to_process := []*os.File{}

	// file, err := os.ReadFile(path)
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// files_to_process = append(files_to_process, file)

	scanner := bufio.NewScanner(file)

	funcPattern := regexp.MustCompile(`(\w+)\s*\(.*{|(\w+)\s*\(\n.*\n\)\s+{`)

	for scanner.Scan() {
		line := scanner.Text()
		// Check if the line contains a function declaration
		if match := funcPattern.FindStringSubmatch(line); match != nil {
			// The first element of match is the full match,
			// and the second element is the function name
			functionName := match[1]
			fmt.Println("Found function:", functionName)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
