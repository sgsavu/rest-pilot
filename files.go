package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func loadTestsFromFile(filename string) ([]Test, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var tests []Test
	err = json.Unmarshal(data, &tests)
	if err != nil {
		return nil, err
	}

	return tests, nil
}

func scanForTestFiles(rootDir string) ([]TestFile, error) {
	var allTestFiles []TestFile
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" && filepath.Base(path)[len(filepath.Base(path))-10:] == ".test.json" {
			tests, err := loadTestsFromFile(path)
			if err != nil {
				return err
			}
			allTestFiles = append(allTestFiles, TestFile{Filename: path, Tests: tests})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return allTestFiles, nil
}
