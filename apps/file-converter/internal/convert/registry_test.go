package convert

import (
	"context"
	"errors"
	"io"
	"reflect"
	"slices"
	"strings"
	"testing"
)

type DumbConverter struct {
	name       string
	mediaTypes map[MediaType][]MediaType
}

func (c *DumbConverter) SupportedFormats() map[MediaType][]MediaType {
	return c.mediaTypes
}

func (c *DumbConverter) Name() string {
	return c.name
}

func (c *DumbConverter) Convert(_ context.Context, _ io.ReadSeeker, _ io.Writer, _ Options) error {
	return nil
}

func NewDumbConverter(name string) *DumbConverter {
	return &DumbConverter{
		name: name,
		mediaTypes: map[MediaType][]MediaType{
			"image/png": {"image/jpeg"},
		},
	}
}
func NewDumbConverterMT(name string, mediaTypes map[MediaType][]MediaType) *DumbConverter {
	return &DumbConverter{
		name:       name,
		mediaTypes: mediaTypes,
	}
}

// DumbLifecycleConverter is a DumbConverter that also implements Lifecycle,
// recording each Start/Stop call (in order) to a shared log so tests can
// assert both invocation and ordering across multiple converters.
type DumbLifecycleConverter struct {
	*DumbConverter
	startErr error
	stopErr  error
	log      *[]string
}

func NewDumbLifecycleConverter(name string, log *[]string, startErr, stopErr error) *DumbLifecycleConverter {
	return &DumbLifecycleConverter{
		DumbConverter: NewDumbConverter(name),
		startErr:      startErr,
		stopErr:       stopErr,
		log:           log,
	}
}

func (c *DumbLifecycleConverter) Start() error {
	*c.log = append(*c.log, "start:"+c.Name())
	return c.startErr
}

func (c *DumbLifecycleConverter) Stop() error {
	*c.log = append(*c.log, "stop:"+c.Name())
	return c.stopErr
}

func TestRegistry_Register(t *testing.T) {
	t.Run("Register simple converter", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(NewDumbConverter("test"))
		registry.Register(NewDumbConverter("test2"))

		converters := registry.List()

		want := []string{"test", "test2"}
		have := make([]string, len(converters))

		for i, conv := range converters {
			have[i] = conv.Name()
		}

		if !reflect.DeepEqual(want, have) {
			t.Errorf("want %v, have %v", want, have)
		}
	})
}

func TestRegistry_Formats(t *testing.T) {
	t.Run("List all registered formats, and lookup converters", func(t *testing.T) {
		registry := NewRegistry()

		conv1 := NewDumbConverterMT("test", map[MediaType][]MediaType{
			"testi": {"kowski"},
		})
		conv2 := NewDumbConverterMT("test", map[MediaType][]MediaType{
			"testi":     {"wakowski", "kowski"},
			"something": {"else"},
		})

		registry.Register(conv1)
		registry.Register(conv2)

		want := map[MediaType][]MediaType{"testi": {"kowski", "wakowski"}, "something": {"else"}}
		wantConvs := map[string]Converter{
			"testi/kowski":   conv1,
			"testi/wakowski": conv2,
			"something/else": conv2,
			"testi/unknown":  nil,
		}

		have := registry.Formats()

		for _, targets := range have {
			slices.Sort(targets)
		}

		if !reflect.DeepEqual(want, have) {
			t.Errorf("want %v, have %v", want, have)
		}

		for typ, conv := range wantConvs {
			split := strings.Split(typ, "/")
			associated, found := registry.Lookup(MediaType(split[0]), MediaType(split[1]))
			if conv != associated || (conv != nil) != found {
				t.Errorf("want %v, have %v [in: %s]", conv, associated, typ)
			}
		}
	})
}

func TestRegistry_StartAll(t *testing.T) {
	t.Run("starts only lifecycle-aware converters, in registration order", func(t *testing.T) {
		registry := NewRegistry()
		var log []string

		registry.Register(NewDumbConverter("plain"))
		registry.Register(NewDumbLifecycleConverter("first", &log, nil, nil))
		registry.Register(NewDumbLifecycleConverter("second", &log, nil, nil))

		if err := registry.StartAll(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []string{"start:first", "start:second"}
		if !slices.Equal(log, want) {
			t.Errorf("want %v, have %v", want, log)
		}
	})

	t.Run("joins errors from all failing converters", func(t *testing.T) {
		registry := NewRegistry()
		var log []string
		errFirst := errors.New("first boom")
		errThird := errors.New("third boom")

		registry.Register(NewDumbLifecycleConverter("first", &log, errFirst, nil))
		registry.Register(NewDumbLifecycleConverter("second", &log, nil, nil))
		registry.Register(NewDumbLifecycleConverter("third", &log, errThird, nil))

		err := registry.StartAll()
		if !errors.Is(err, errFirst) || !errors.Is(err, errThird) {
			t.Errorf("want joined error containing %v and %v, have %v", errFirst, errThird, err)
		}
	})
}

func TestRegistry_StopAll(t *testing.T) {
	t.Run("stops only lifecycle-aware converters, in reverse registration order", func(t *testing.T) {
		registry := NewRegistry()
		var log []string

		registry.Register(NewDumbLifecycleConverter("first", &log, nil, nil))
		registry.Register(NewDumbConverter("plain"))
		registry.Register(NewDumbLifecycleConverter("second", &log, nil, nil))

		if err := registry.StopAll(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []string{"stop:second", "stop:first"}
		if !slices.Equal(log, want) {
			t.Errorf("want %v, have %v", want, log)
		}
	})

	t.Run("joins errors from all failing converters", func(t *testing.T) {
		registry := NewRegistry()
		var log []string
		errFirst := errors.New("first boom")
		errSecond := errors.New("second boom")

		registry.Register(NewDumbLifecycleConverter("first", &log, nil, errFirst))
		registry.Register(NewDumbLifecycleConverter("second", &log, nil, errSecond))

		err := registry.StopAll()
		if !errors.Is(err, errFirst) || !errors.Is(err, errSecond) {
			t.Errorf("want joined error containing %v and %v, have %v", errFirst, errSecond, err)
		}
	})
}
