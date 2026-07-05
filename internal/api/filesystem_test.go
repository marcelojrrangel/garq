package api

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

func TestListRoots(t *testing.T) {
	a := &API{}
	roots, err := a.ListRoots()
	if err != nil {
		t.Fatalf("ListRoots failed: %v", err)
	}
	fmt.Printf("GOOS: %s\n", runtime.GOOS)
	fmt.Printf("Roots found: %d\n", len(roots))
	for _, r := range roots {
		fmt.Printf("  Drive: %s\n", r)
		info, err := os.Stat(r)
		if err != nil {
			fmt.Printf("    Stat error: %v\n", err)
		} else {
			fmt.Printf("    IsDir: %v, Mode: %s\n", info.IsDir(), info.Mode())
		}
	}
	if len(roots) == 0 && runtime.GOOS == "windows" {
		t.Fatal("Expected at least one drive on Windows")
	}
}
