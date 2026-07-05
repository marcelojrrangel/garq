//go:build !windows

package main

import "os"

// RecycleItems em plataformas não-Windows faz exclusão permanente como fallback.
func RecycleItems(paths []string) error {
	for _, p := range paths {
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}
	return nil
}
