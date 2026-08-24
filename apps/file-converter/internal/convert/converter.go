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
	Target MediaType
}

// Converter A file converter
type Converter interface {
	// Name The (unique) name of this converter, mostly for debugging purposes
	Name() string
	// SupportedTargets A list of supported output media types this converter can handle for the given input media type.
	SupportedTargets(src MediaType) []MediaType
	// Convert The actual conversion
	Convert(ctx context.Context, in io.ReadSeeker, out io.Writer, opts Options) error
}
