package convert

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

var types = []MediaType{"image/avif", "image/gif", "image/jpeg", "image/png", "image/webp"}

type ImageConverter struct {
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

func (c ImageConverter) SupportedTargets(src MediaType) []MediaType {
	if !slices.Contains(types, src) {
		return nil
	}

	var targets []MediaType
	for _, t := range types {
		if t != src {
			targets = append(targets, t)
		}
	}
	return targets
}

func (c ImageConverter) Convert(_ context.Context, in io.ReadSeeker, out io.Writer, opts Options) error {
	image, err := vips.NewImageFromReader(in)
	if err != nil {
		return err
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
		return err
	}

	_, err = out.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

func NewImageConverter() *ImageConverter {
	return &ImageConverter{}
}
