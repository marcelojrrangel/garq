package api

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func (a *API) ListRoots() ([]string, error) {
	fmt.Println("ListRoots called")
	if runtime.GOOS != "windows" {
		fmt.Println("Not windows, returning /")
		return []string{"/"}, nil
	}
	roots := make([]string, 0, 8)
	for c := 'A'; c <= 'Z'; c++ {
		drive := string(c) + ":\\"
		fmt.Printf("Checking drive %s...\n", drive)
		if _, err := os.Stat(drive); err == nil {
			roots = append(roots, drive)
			fmt.Printf("  -> found %s\n", drive)
		}
	}
	fmt.Printf("ListRoots returning %d drives\n", len(roots))
	return roots, nil
}

func (a *API) ListDirectory(path string) ([]DirEntryDTO, error) {
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return nil, errors.New("invalid path")
	}
	if !filepath.IsAbs(cleanPath) {
		return nil, errors.New("path must be absolute")
	}
	entries, err := os.ReadDir(cleanPath)
	if err != nil {
		return nil, err
	}

	items := make([]DirEntryDTO, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		fullPath := filepath.Join(path, e.Name())
		items = append(items, DirEntryDTO{
			Name:    e.Name(),
			Path:    fullPath,
			IsDir:   e.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
			Mode:    info.Mode().String(),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	return items, nil
}
