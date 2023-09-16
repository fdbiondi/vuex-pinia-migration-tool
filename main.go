package main

import (
	"fileutil"
	"fmt"
	"log"
	"os"
	"tool"
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

	fmt.Printf("source path '%s'", srcPath)

	tool.Transform(outputPath)
}