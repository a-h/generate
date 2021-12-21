package utils

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func FileNameCreation(fileName string) string {
	return fmt.Sprintf("%s%s", fileName[:len(fileName)-len(filepath.Ext(fileName))], ".go")
}

// ReadFiles Reads file or files From Directories
func ReadFiles(inputPath string) ([]string, error) {
	stat, err := os.Stat(inputPath)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		//single file entry
		fullPath, err := filepath.Abs(inputPath)
		if err != nil {
			return nil, err
		}

		return []string{fullPath}, nil
	}

	//read a directory
	files, err := os.ReadDir(inputPath)
	if err != nil {
		return nil, err
	}

	inputPath, err = filepath.Abs(inputPath)
	if err != nil {
		return nil, err
	}

	filePaths := make([]string, len(files))
	for i, file := range files {
		filePaths[i] = filepath.Join(inputPath, file.Name())
	}

	return filePaths, nil
}

// PackageFormat formatting the package name and the Directory Name
func PackageFormat(outputDir string, file string) (packageDirectory string, packageName string) {
	parts := strings.Split(filepath.Base(file), ".")
	packageDirectory = path.Join(outputDir, parts[0])
	packageName = "models"
	return
}

func ParseFlags() (string, string) {
	inputDir := flag.String("input", "../schemas", "Please Enter The Input Directory")
	outputDir := flag.String("output", "../output", "Please Enter The Input Directory")
	flag.Parse()
	return *inputDir, *outputDir
}
