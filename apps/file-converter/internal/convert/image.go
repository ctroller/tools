package convert

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

type VipsConverter struct {
}

func (c VipsConverter) Start() error {
	return vips.Startup(nil)
}

func (c VipsConverter) Stop() error {
	vips.Shutdown()
	return nil
}

func (c VipsConverter) Name() string {
	return "VipsConverter"
}

func (c VipsConverter) CanHandle(src, tgt MediaType) bool {
	return src != tgt && strings.HasPrefix(string(src), "image/") && strings.HasPrefix(string(tgt), "image/")
}

func (c VipsConverter) Convert(_ context.Context, in io.ReadSeeker, out io.Writer, opts Options) error {
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

func NewVipsConverter() *VipsConverter {
	return &VipsConverter{}
}
