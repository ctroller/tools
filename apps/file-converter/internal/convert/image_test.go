package convert

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

const imgJPEG = "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAAMCAgMCAgMDAwMEAwMEBQgFBQQEBQoHBwYIDAoMDAsKCwsNDhIQDQ4RDgsLEBYQERMUFRUVDA8XGBYUGBIUFRT/wAALCAABAAEBAREA/8QAFAABAAAAAAAAAAAAAAAAAAAACf/EABQQAQAAAAAAAAAAAAAAAAAAAAD/2gAIAQEAAD8AKp//2Q=="
const imgPNG = "iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAApgAAAKYB3X3/OAAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAANCSURBVEiJtZZPbBtFFMZ/M7ubXdtdb1xSFyeilBapySVU8h8OoFaooFSqiihIVIpQBKci6KEg9Q6H9kovIHoCIVQJJCKE1ENFjnAgcaSGC6rEnxBwA04Tx43t2FnvDAfjkNibxgHxnWb2e/u992bee7tCa00YFsffekFY+nUzFtjW0LrvjRXrCDIAaPLlW0nHL0SsZtVoaF98mLrx3pdhOqLtYPHChahZcYYO7KvPFxvRl5XPp1sN3adWiD1ZAqD6XYK1b/dvE5IWryTt2udLFedwc1+9kLp+vbbpoDh+6TklxBeAi9TL0taeWpdmZzQDry0AcO+jQ12RyohqqoYoo8RDwJrU+qXkjWtfi8Xxt58BdQuwQs9qC/afLwCw8tnQbqYAPsgxE1S6F3EAIXux2oQFKm0ihMsOF71dHYx+f3NND68ghCu1YIoePPQN1pGRABkJ6Bus96CutRZMydTl+TvuiRW1m3n0eDl0vRPcEysqdXn+jsQPsrHMquGeXEaY4Yk4wxWcY5V/9scqOMOVUFthatyTy8QyqwZ+kDURKoMWxNKr2EeqVKcTNOajqKoBgOE28U4tdQl5p5bwCw7BWquaZSzAPlwjlithJtp3pTImSqQRrb2Z8PHGigD4RZuNX6JYj6wj7O4TFLbCO/Mn/m8R+h6rYSUb3ekokRY6f/YukArN979jcW+V/S8g0eT/N3VN3kTqWbQ428m9/8k0P/1aIhF36PccEl6EhOcAUCrXKZXXWS3XKd2vc/TRBG9O5ELC17MmWubD2nKhUKZa26Ba2+D3P+4/MNCFwg59oWVeYhkzgN/JDR8deKBoD7Y+ljEjGZ0sosXVTvbc6RHirr2reNy1OXd6pJsQ+gqjk8VWFYmHrwBzW/n+uMPFiRwHB2I7ih8ciHFxIkd/3Omk5tCDV1t+2nNu5sxxpDFNx+huNhVT3/zMDz8usXC3ddaHBj1GHj/As08fwTS7Kt1HBTmyN29vdwAw+/wbwLVOJ3uAD1wi/dUH7Qei66PfyuRj4Ik9is+hglfbkbfR3cnZm7chlUWLdwmprtCohX4HUtlOcQjLYCu+fzGJH2QRKvP3UNz8bWk1qMxjGTOMThZ3kvgLI5AzFfo379UAAAAASUVORK5CYII="

func TestImageConverter_Formats(t *testing.T) {
	t.Run("Should return all supported formats", func(t *testing.T) {
		conv := NewImageConverter()
		want := make(map[MediaType][]MediaType, len(imageTypes))

		for _, src := range imageTypes {
			want[src] = make([]MediaType, len(imageTypes))
			copy(want[src], imageTypes)

			want[src] = slices.DeleteFunc(want[src], func(tgt MediaType) bool {
				return tgt == src
			})
		}

		if !reflect.DeepEqual(want, conv.SupportedFormats()) {
			out, _ := json.MarshalIndent(want, "", "\t")
			t.Errorf("SupportedFormats() = %v, want %v", conv.SupportedFormats(), string(out))
		}
	})
}

var formatToVips = map[MediaType]vips.ImageType{
	"image/jpeg": vips.ImageTypeJPEG,
	"image/png":  vips.ImageTypePNG,
	"image/avif": vips.ImageTypeAVIF,
	"image/webp": vips.ImageTypeWEBP,
	"image/gif":  vips.ImageTypeGIF,
}

func TestImageConverter_Convert(t *testing.T) {
	type Test struct {
		in    MediaType
		image string
	}

	tests := []Test{
		{
			in:    "image/jpeg",
			image: imgJPEG,
		},
		{
			in:    "image/png",
			image: imgPNG,
		},
	}

	conv := NewImageConverter()
	for _, test := range tests {
		image, err := base64.StdEncoding.DecodeString(test.image)
		if err != nil {
			t.Errorf("failed to decode image: %v", err)
			return
		}

		for _, format := range imageTypes {
			if format != test.in {

				t.Run("Convert "+string(test.in)+" to "+string(format), func(t *testing.T) {
					in := bytes.NewReader(image)
					out := bytes.NewBuffer(nil)

					if err := conv.Convert(context.Background(), in, out, Options{
						Target: format,
					}); err != nil {
						t.Errorf("Convert() = %v", err)
						return
					}

					if len(out.Bytes()) == 0 {
						t.Errorf("output is empty")
						return
					}

					img, err := vips.NewImageFromBuffer(out.Bytes())
					if err != nil {
						t.Errorf("failed to create image from buffer: %v", err)
						return
					}

					if img.Format() != formatToVips[format] {
						t.Errorf("unexpected image format: got %v, want %v", img.Format(), formatToVips[format])
						return
					}
				})
			}
		}
	}
}
