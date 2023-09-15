package tool

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Migrate(dest_dir string) {
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
			readAndTransform(files_to_process, dest_dir)

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

func readAndTransform(files []string, dest_dir string) {
	fmt.Printf("directory output '%s'\n", dest_dir)
	fmt.Println("")

	// files_to_process := []*os.File{}
	// runtime := goja.New()

	for _, origin_filepath := range files {
		fmt.Println("filename => ", origin_filepath)
		fmt.Println("------------------------")

		file, err := os.Open(origin_filepath)
		if err != nil {
			log.Fatal(err)
		}

		// _, err = runtime.RunString(string(file))
		// if err != nil {
		// 	fmt.Println("Error evaluating JavaScript code:", err)
		// 	return
		// }

		defer file.Close()
		// files_to_process = append(files_to_process, file)

		// scanner := bufio.NewScanner(file)
		// for scanner.Scan() {
		// fmt.Println(scanner.Text())
		// }

		// if err := scanner.Err(); err != nil {
		// 	log.Fatal(err)
		// }

		fmt.Println("------------------------")
		fmt.Println("")
	}
}
