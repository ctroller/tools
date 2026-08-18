package convert

import (
	"context"
	"io"
)

type MediaType string

// Lifecycle interface for converters that need custom handling (allocating / freeing resources, config, ...)
type Lifecycle interface {
	Start() error
	Stop() error
}

type Options struct {
}

type Converter interface {
	CanHandle(src, tgt MediaType) bool
	Convert(ctx context.Context, in io.ReadSeeker, out io.Writer, opts Options) error
}
