package converter

import (
	"io"
	"os"
	"path/filepath"

	"github.com/azarc-io/json-schema-to-go-struct-generator/pkg/inputs"
	"github.com/azarc-io/json-schema-to-go-struct-generator/pkg/utils"
	"github.com/pkg/errors"
)

func Convert(inputFiles []string, outputDir string) error {
	schemas, err := inputs.ReadInputFiles(inputFiles, false) // passing true will check for schema key in the file
	if err != nil {
		return errors.Wrapf(err, "error while reading input file")

	}
	generatorInstance := inputs.New(schemas...) // instance of generator which will produce structs
	err = generatorInstance.CreateTypes()
	if err != nil {
		return errors.Wrapf(err, "error while generating instance for  proudcing structs")

	}

	for _, file := range inputFiles {
		var w io.Writer
		packageDirectory, packageName := utils.PackageFormat(outputDir, file)

		err = os.MkdirAll(packageDirectory, 0755)
		if err != nil {
			return errors.Wrapf(err, "error while creating directory")

		}
		w, err = os.Create(filepath.Join(packageDirectory, filepath.Base(utils.FileNameCreation(file))))

		if err != nil {
			return errors.Wrapf(err, "error while generating Files")

		}

		// Model Generation Method Called
		wd, _ := os.Getwd()
		relativePath, err := filepath.Rel(wd, file)
		if err != nil {
			return err
		}

		inputs.Output(w, generatorInstance, packageName, relativePath)

	}
	return nil
}
