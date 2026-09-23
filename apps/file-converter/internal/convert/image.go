package convert

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/dustin/go-humanize"
	"trox.dev/file-converter/internal/common"
)

var imageTypes = []MediaType{MediaTypeAVIF, MediaTypeGIF, MediaTypeJPEG, MediaTypePNG, MediaTypeWEBP}

type ImageConverter struct {
	formats         map[MediaType][]MediaType
	maxImagePxCount int64
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

	height := image.Height()
	width := image.Width()
	px := int64(width) * int64(height)

	if px > c.maxImagePxCount {
		return common.IllegalArgErr{Msg: "image pixel count is too large: " + humanize.Comma(c.maxImagePxCount) + "px maximum, got " + humanize.Comma(px) + "px"}
	}

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

func NewImageConverter(maxImagePxCount int64) *ImageConverter {
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
		maxImagePxCount,
	}
}
