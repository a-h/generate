package main

import (
	"fmt"
	"github.com/azarc-io/json-schema-to-go-struct-generator/pkg/converter"
	"github.com/azarc-io/json-schema-to-go-struct-generator/pkg/utils"
)

func main() {
	inputPath, outputDir := utils.ParseFlags() // Parsing the cl flags
	files, err := utils.ReadFiles(inputPath)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Processing files: %v\n", files)
	err = converter.Convert(files, outputDir)
	if err != nil {
		panic(err)
	}
}
