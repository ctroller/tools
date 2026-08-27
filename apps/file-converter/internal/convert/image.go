package convert

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

var imageTypes = []MediaType{"image/avif", "image/gif", "image/jpeg", "image/png", "image/webp"}

type ImageConverter struct {
	formats map[MediaType][]MediaType
}

func (c ImageConverter) Start() error {
	return vips.Startup(nil)
}

func (c ImageConverter) Stop() error {
	vips.Shutdown()
	return nil
}

func (c ImageConverter) Name() string {
	return "ImageConverter"
}

func (c ImageConverter) SupportedFormats() map[MediaType][]MediaType {
	return c.formats
}

func (c ImageConverter) Convert(_ context.Context, in io.ReadSeeker, out io.Writer, opts Options) error {
	image, err := vips.NewImageFromReader(in)
	if err != nil {
		return fmt.Errorf("can't create image from reader: %w", err)
	}
	defer image.Close()

	return internalConvert(image, out, opts)
}

func internalConvert(image *vips.ImageRef, out io.Writer, opts Options) error {
	tgt := strings.TrimPrefix(string(opts.Target), "image/")
	var buf []byte
	var err error

	slog.Info("Converting image ", "from", image.Format(), "to", opts.Target)

	switch tgt {
	case "jpg", "jpeg":
		buf, _, err = image.ExportJpeg(nil)
	case "avif":
		buf, _, err = image.ExportAvif(nil)
	case "webp":
		buf, _, err = image.ExportWebp(nil)
	case "png":
		buf, _, err = image.ExportPng(nil)
	case "gif":
		buf, _, err = image.ExportGIF(nil)
	default:
		return fmt.Errorf("unsupported target format %s", opts.Target)
	}

	if err != nil {
		return fmt.Errorf("can't create export image: %w", err)
	}

	_, err = out.Write(buf)
	if err != nil {
		return fmt.Errorf("unable to write to output: %w", err)
	}

	return nil
}

func NewImageConverter() *ImageConverter {
	var formats = make(map[MediaType][]MediaType)
	for _, src := range imageTypes {
		var tmpFormats []MediaType
		for _, tgt := range imageTypes {
			if src != tgt {
				tmpFormats = append(tmpFormats, tgt)
			}
		}

		formats[src] = tmpFormats
	}

	return &ImageConverter{
		formats,
	}
}
