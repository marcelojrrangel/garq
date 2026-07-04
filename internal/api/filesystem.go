package api

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type fsItem struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime string
	Mode    string
}

func (a *API) ListRoots() ([]string, error) {
	if runtime.GOOS != "windows" {
		return []string{"/"}, nil
	}
	roots := make([]string, 0, 8)
	for c := 'A'; c <= 'Z'; c++ {
		drive := string(c) + ":\\"
		if _, err := os.Stat(drive); err == nil {
			roots = append(roots, drive)
		}
	}
	return roots, nil
}

func (a *API) ListDirectory(path string) ([]map[string]any, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	items := make([]fsItem, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		fullPath := filepath.Join(path, e.Name())
		items = append(items, fsItem{
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

	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		out = append(out, map[string]any{
			"name":     it.Name,
			"path":     it.Path,
			"is_dir":   it.IsDir,
			"size":     it.Size,
			"mod_time": it.ModTime,
			"mode":     it.Mode,
		})
	}
	return out, nil
}
