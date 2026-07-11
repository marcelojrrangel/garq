package copy

// CopierAdapter satisfies worker.FileCopier using the platform-specific CopyFile.
type CopierAdapter struct{}

// CopyFile copies src to destDir using the platform-specific implementation.
func (CopierAdapter) CopyFile(src, destDir string) error {
	return CopyFile(src, destDir)
}
