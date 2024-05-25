package main

import (
	"log"
	"os"
	"path/filepath"
)

var DIR string = ""

func getFilepath(filename string) string {
	if DIR == "" {

		if len(os.Args) != 3 {
			log.Fatal("Not enough args")
		}
		dir, err := filepath.Abs(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		DIR = dir
	}
	return filepath.Join(DIR, filename)
}

func readFile(filename string) ([]byte, bool) {
	path := getFilepath(filename)
	file, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, true
	}
	if err != nil {
		log.Fatal(err)
	}
	return file, false
}

func writeFile(filename string, data []byte) {
	path := getFilepath(filename)
	err := os.WriteFile(path, data, 0755)
	if err != nil {
		log.Fatal(err)
	}
}
