//go:build !windows

package shell

import "errors"

// RecycleItems is a no-op stub on non-Windows platforms.
func RecycleItems(paths []string) error {
	return errors.New("recycle bin is only available on Windows")
}
