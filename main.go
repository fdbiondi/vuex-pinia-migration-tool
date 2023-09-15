package main

import (
	"fileutil"
	"fmt"
	"log"
	"os"
	"tool"
)

func main() {
	output_path := "./tmp/store-migrated/"
	src_path := "./tmp/store/"

	err := os.RemoveAll(output_path)
	if err != nil {
		log.Fatal(err)
	}

	err = fileutil.CopyDirectory(src_path, output_path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("source path '%s'", src_path)

	tool.Transform(output_path)
}
