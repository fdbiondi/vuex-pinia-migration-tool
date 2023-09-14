package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	output_path := "./tmp/store-migrated/"
	root_path := "./tmp/store/"
	current_dir := ""
	current_path := ""
	files_to_process := []string{}

	err := os.RemoveAll(output_path)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir(output_path, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(root_path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Err: ", err)
			return err
		}

		if root_path == path {
			return nil
		}

		if !strings.Contains(path, current_path) {
			process(files_to_process, root_path, output_path)

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

		return nil
	})

	if err != nil {
		fmt.Println("Err: ", err)
	}

	fmt.Println("origin path: ", root_path)
	fmt.Println("output path: ", output_path)
}

func process(files []string, root_path string, output_path string) {
	fmt.Println("")
	fmt.Println("files: ")

	root_path = strings.Replace(root_path, "./", "", 1)
	output_path = strings.Replace(output_path, "./", "", 1)

	for _, file := range files {
		new_path := strings.Replace(file, root_path, output_path, 1)

		fmt.Println(file, " -> ", new_path)

		if _, err := os.Stat(new_path); os.IsNotExist(err) {
			os.MkdirAll(filepath.Dir(new_path), 0700)
		}

		file, err := os.Create(new_path)
		if err != nil {
			log.Fatal(err)
		}

		file.Write([]byte("test"))
		file.Close()

	}

}
