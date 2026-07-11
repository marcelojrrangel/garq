package compress

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func findTest7z(t *testing.T) string {
	t.Helper()
	if custom := os.Getenv("GARQ_7Z_PATH"); custom != "" {
		if _, err := os.Stat(custom); err == nil {
			return custom
		}
	}
	candidates := []string{
		"../7z/7z.exe",
		"../7z/7zz",
		"../../7z/7z.exe",
		"../../7z/7zz",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(p)
			t.Setenv("GARQ_7Z_PATH", abs)
			return abs
		}
	}
	t.Skip("7z not found, skipping compress tests")
	return ""
}

func TestCompressAndExtract(t *testing.T) {
	_ = findTest7z(t)
	srcDir := t.TempDir()
	outDir := t.TempDir()
	extractDir := t.TempDir()

	files := map[string]string{
		"a.txt": "content A",
		"b.txt": "content B",
	}
	for name, content := range files {
		os.WriteFile(filepath.Join(srcDir, name), []byte(content), 0644)
	}

	archive := filepath.Join(outDir, "test.7z")
	sources := []string{
		filepath.Join(srcDir, "a.txt"),
		filepath.Join(srcDir, "b.txt"),
	}

	if err := CompressMany(sources, archive); err != nil {
		t.Fatalf("CompressMany failed: %v", err)
	}

	if _, err := os.Stat(archive); os.IsNotExist(err) {
		t.Fatal("archive was not created")
	}

	if err := Extract(archive, extractDir); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	for name, expected := range files {
		got, err := os.ReadFile(filepath.Join(extractDir, name))
		if err != nil {
			t.Fatalf("read extracted file %s failed: %v", name, err)
		}
		if string(got) != expected {
			t.Errorf("file %s: got %q, want %q", name, got, expected)
		}
	}
}

func TestCompressEmpty(t *testing.T) {
	outDir := t.TempDir()
	archive := filepath.Join(outDir, "empty.7z")

	err := CompressMany([]string{}, archive)
	if err == nil {
		t.Fatal("expected error for empty sources")
	}
}

func TestCompressSingle(t *testing.T) {
	_ = findTest7z(t)
	srcDir := t.TempDir()
	outDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "single.txt")
	os.WriteFile(srcPath, []byte("single file"), 0644)

	archive := filepath.Join(outDir, "single.7z")
	if err := Compress(srcPath, archive); err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	if _, err := os.Stat(archive); os.IsNotExist(err) {
		t.Fatal("archive was not created")
	}
}

func TestExtractToNewDir(t *testing.T) {
	_ = findTest7z(t)
	srcDir := t.TempDir()
	outDir := t.TempDir()
	extractDir := t.TempDir() + "/nested"

	srcPath := filepath.Join(srcDir, "file.txt")
	os.WriteFile(srcPath, []byte("test"), 0644)

	archive := filepath.Join(outDir, "test.7z")
	CompressMany([]string{srcPath}, archive)

	if err := Extract(archive, extractDir); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(extractDir, "file.txt"))
	if err != nil {
		t.Fatalf("read extracted file failed: %v", err)
	}
	if string(got) != "test" {
		t.Errorf("got %q, want %q", got, "test")
	}
}

func TestParseProgress(t *testing.T) {
	tests := []struct {
		line     string
		expected float64
		ok       bool
	}{
		{"\r35%", 0.35, true},
		{"35%", 0.35, true},
		{"\r78% - file.txt", 0.78, true},
		{"\r100%", 1.0, true},
		{"0%", 0.0, true},
		{"100%", 1.0, true},
		{"Extracting  file.txt", 0, false},
		{"", 0, false},
		{"some random text", 0, false},
	}
	for _, tc := range tests {
		got, ok := parseProgress(tc.line)
		if ok != tc.ok || got != tc.expected {
			t.Errorf("parseProgress(%q) = (%v, %v), want (%v, %v)", tc.line, got, ok, tc.expected, tc.ok)
		}
	}
}

func TestCompressManyCtx(t *testing.T) {
	_ = findTest7z(t)
	srcDir := t.TempDir()
	outDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test_data.bin")
	data := make([]byte, 5<<20)
	for i := range data {
		data[i] = byte(i)
	}
	os.WriteFile(srcPath, data, 0644)

	archive := filepath.Join(outDir, "test.7z")
	var progresses []float64
	cb := func(pct float64) {
		progresses = append(progresses, pct)
	}

	ctx := context.Background()
	if err := CompressManyCtx(ctx, []string{srcPath}, archive, cb); err != nil {
		t.Fatalf("CompressManyCtx failed: %v", err)
	}
	if _, err := os.Stat(archive); os.IsNotExist(err) {
		t.Fatal("archive was not created")
	}
	if len(progresses) == 0 {
		t.Log("no intermediate progress reported (small file)")
	} else {
		last := progresses[len(progresses)-1]
		if last < 0.9 {
			t.Errorf("last progress %.2f, expected near 1.0", last)
		}
		t.Logf("progress updates: %d, last: %.2f", len(progresses), last)
	}
}

func TestExtractCtx(t *testing.T) {
	_ = findTest7z(t)
	srcDir := t.TempDir()
	outDir := t.TempDir()
	extractDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "file.txt")
	os.WriteFile(srcPath, []byte("hello world"), 0644)

	archive := filepath.Join(outDir, "test.7z")
	CompressMany([]string{srcPath}, archive)

	var progresses []float64
	cb := func(pct float64) {
		progresses = append(progresses, pct)
	}

	ctx := context.Background()
	if err := ExtractCtx(ctx, archive, extractDir, cb); err != nil {
		t.Fatalf("ExtractCtx failed: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(extractDir, "file.txt"))
	if err != nil {
		t.Fatalf("read extracted file failed: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
	t.Logf("extract progress updates: %d", len(progresses))
}

func TestCompressManyCtxCancel(t *testing.T) {
	_ = findTest7z(t)
	srcDir := t.TempDir()
	outDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "large.bin")
	data := make([]byte, 50<<20)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}
	os.WriteFile(srcPath, data, 0644)

	archive := filepath.Join(outDir, "test.7z")
	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- CompressManyCtx(ctx, []string{srcPath}, archive, nil)
	}()

	<-time.After(200 * time.Millisecond)
	cancel()

	err := <-errCh
	if err == nil {
		t.Fatal("expected error after cancel, got nil")
	}
	if err != context.Canceled {
		t.Logf("got error: %v (may be process kill)", err)
	}
}
