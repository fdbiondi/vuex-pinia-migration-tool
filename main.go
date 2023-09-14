package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root_path := "./tmp/store/"
	current_dir := ""
	current_path := ""
	files_to_process := []string{}

	err := filepath.Walk(root_path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Err: ", err)
			return err
		}

		if root_path == path {
			return nil
		}

		if !strings.Contains(path, current_path) {
			process(files_to_process)

			files_to_process = []string{}
			fmt.Println("")
		}

		if info.IsDir() {
			current_dir = strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
			current_path = path
			fmt.Println("")
			fmt.Println("current_dir: ", strings.ToUpper(current_dir))

			return nil
		}

		files_to_process = append(files_to_process, path)

		fmt.Printf("dir: %v, path: %s, name: %s\n", info.IsDir(), path, filepath.Base(path))

		return nil
	})

	if err != nil {
		fmt.Println("Err: ", err)
	}
}

func process(files []string) {
	fmt.Println("")
	fmt.Println("files: ", files)
	for _, file := range files {
		fmt.Println(file)
	}
}
