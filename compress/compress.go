package compress

import "context"

// CLIAdapter satisfies worker.Compressor using the 7z CLI implementation.
type CLIAdapter struct{}

// CompressManyCtx compresses sources into dest with progress reporting.
func (CLIAdapter) CompressManyCtx(ctx context.Context, sources []string, dest string, progressCb func(float64)) error {
	return CompressManyCtx(ctx, sources, dest, progressCb)
}

// ExtractCtx extracts archive into dest with progress reporting.
func (CLIAdapter) ExtractCtx(ctx context.Context, archive, dest string, progressCb func(float64)) error {
	return ExtractCtx(ctx, archive, dest, progressCb)
}
