package parser

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
)

var getterPattern = map[string]*regexp.Regexp{
	// "object": regexp.MustCompile(`^\s{2}\w+\s{2}\(\s{2}\w+\s{2}\)\s{2}\{`),
}

func parseGetters(filesMap map[string]*os.File) []string {
	file, ok := filesMap["getters"]
	if !ok {
		return []string{}
	}

	fmt.Printf("parsing: %s\n\n", file.Name())
	scanner := bufio.NewScanner(file)

	var lines []string

	// for scanner.Scan() {
	// 	line := scanner.Text()
	//
	// }

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines
}
