package main

import (
	"fileutil"
	"fmt"
	"log"
	"os"
	"parser"
)

func main() {
	outputPath := "./tmp/store-migrated/"
	srcPath := "./tmp/store/"

	err := os.RemoveAll(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	err = fileutil.CopyDirectory(srcPath, outputPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("source path '%s'\n", srcPath)
	fmt.Printf("output path '%s'\n\n", outputPath)

	parser.Transform(outputPath)
}
