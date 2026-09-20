package convert

import (
	"log/slog"
	"mime"
)

type MediaType string

const (
	MediaTypeJPEG MediaType = "image/jpeg"
	MediaTypePNG  MediaType = "image/png"
	MediaTypeGIF  MediaType = "image/gif"
	MediaTypeAVIF MediaType = "image/avif"
	MediaTypeWEBP MediaType = "image/webp"
)

const (
	ExtJPEG = ".jpeg"
	ExtPNG  = ".png"
	ExtGIF  = ".gif"
	ExtAVIF = ".avif"
	ExtWEBP = ".webp"
)

var extMappings = map[MediaType]string{
	MediaTypeJPEG: ExtJPEG,
	MediaTypePNG:  ExtPNG,
	MediaTypeGIF:  ExtGIF,
	MediaTypeAVIF: ExtAVIF,
	MediaTypeWEBP: ExtWEBP,
}

func (m MediaType) Ext() string {
	ext, found := extMappings[m]
	if !found {
		ext = ""
		extensions, err := mime.ExtensionsByType(string(m))
		if err != nil {
			slog.Error("failed to get mime extensions", "err", err)
		} else if len(extensions) > 0 {
			ext = extensions[0]
		}
	}

	return ext
}
