package main

import "os"

func countDirContents(path string) (files, folders int) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			folders++
		} else {
			files++
		}
	}
	return
}
