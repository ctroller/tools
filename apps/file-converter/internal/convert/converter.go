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
//
// Name The (unique) name of this converter, mostly for debugging purposes
// SupportedFormats A map of supported output media types this converter can handle for the given input media type (=key).
// Convert The actual conversion method, reading from `in` into `out`
type Converter interface {
	Name() string
	SupportedFormats() map[MediaType][]MediaType
	Convert(ctx context.Context, in io.ReadSeeker, out io.Writer, opts Options) error
}
