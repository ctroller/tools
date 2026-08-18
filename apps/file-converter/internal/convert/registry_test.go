package convert_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"trox.dev/file-converter/internal/convert"
)

// fakeConverter is a minimal Converter test double. A nil canHandle behaves
// as "matches nothing" so it's safe to leave unset in tests that don't care.
type fakeConverter struct {
	canHandle func(src, tgt convert.MediaType) bool
}

func (f *fakeConverter) Name() string {
	return "fakeConverter"
}

func (f *fakeConverter) CanHandle(src, tgt convert.MediaType) bool {
	if f.canHandle == nil {
		return false
	}
	return f.canHandle(src, tgt)
}

func (f *fakeConverter) Convert(context.Context, io.ReadSeeker, io.Writer, convert.Options) error {
	return nil
}

type fakeLifecycleConverter struct {
	fakeConverter
	name     string
	startErr error
	stopErr  error
	started  bool
	stopped  bool
	order    *[]string // if set, Start/Stop append name here to record call order
}

func (f *fakeLifecycleConverter) Name() string {
	return "fakeLifecycleConverter " + f.name
}

func (f *fakeLifecycleConverter) Start() error {
	f.started = true
	if f.order != nil {
		*f.order = append(*f.order, f.name)
	}
	return f.startErr
}

func (f *fakeLifecycleConverter) Stop() error {
	f.stopped = true
	if f.order != nil {
		*f.order = append(*f.order, f.name)
	}
	return f.stopErr
}

func TestRegistry_Lookup(t *testing.T) {
	pngToWebp := &fakeConverter{canHandle: func(src, tgt convert.MediaType) bool {
		return src == "image/png" && tgt == "image/webp"
	}}
	anyImage := &fakeConverter{canHandle: func(src, tgt convert.MediaType) bool {
		return strings.HasPrefix(string(src), "image/") &&
			strings.HasPrefix(string(tgt), "image/") &&
			src != tgt
	}}

	tests := []struct {
		name      string
		register  []convert.Converter
		src, tgt  convert.MediaType
		want      convert.Converter
		wantNoHit bool
	}{
		{
			name:     "specialized transformer wins when registered first",
			register: []convert.Converter{pngToWebp, anyImage},
			src:      "image/png",
			tgt:      "image/webp",
			want:     pngToWebp,
		},
		{
			name:     "generic transformer catches pairs the specialized one doesn't claim",
			register: []convert.Converter{pngToWebp, anyImage},
			src:      "image/avif",
			tgt:      "image/jpeg",
			want:     anyImage,
		},
		{
			name:      "no registered transformer matches",
			register:  []convert.Converter{pngToWebp},
			src:       "image/avif",
			tgt:       "image/jpeg",
			wantNoHit: true,
		},
		{
			name:      "identical source and target does not match, even against a wildcard transformer",
			register:  []convert.Converter{anyImage},
			src:       "image/png",
			tgt:       "image/png",
			wantNoHit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &convert.Registry{}
			for _, c := range tt.register {
				r.Register(c)
			}

			got, ok := r.Lookup(tt.src, tt.tgt)

			if tt.wantNoHit {
				if ok {
					t.Fatalf("Lookup(%q, %q) = %v, true; want ok=false", tt.src, tt.tgt, got)
				}
				return
			}
			if !ok {
				t.Fatalf("Lookup(%q, %q) ok = false; want true", tt.src, tt.tgt)
			}
			if got != tt.want {
				t.Fatalf("Lookup(%q, %q) returned an unexpected transformer", tt.src, tt.tgt)
			}
		})
	}
}

func TestRegistry_Formats(t *testing.T) {
	known := []convert.MediaType{"image/png", "image/webp", "image/avif"}

	t.Run("every pair a registered transformer can handle appears in the matrix", func(t *testing.T) {
		r := &convert.Registry{}
		r.Register(&fakeConverter{canHandle: func(src, tgt convert.MediaType) bool {
			return strings.HasPrefix(string(src), "image/") &&
				strings.HasPrefix(string(tgt), "image/") &&
				src != tgt
		}})

		got := r.Formats(known)
		want := map[convert.MediaType][]convert.MediaType{
			"image/png":  {"image/webp", "image/avif"},
			"image/webp": {"image/png", "image/avif"},
			"image/avif": {"image/png", "image/webp"},
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Formats() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("a format with no valid target is absent from the map, not present with an empty slice", func(t *testing.T) {
		r := &convert.Registry{}
		r.Register(&fakeConverter{canHandle: func(src, tgt convert.MediaType) bool {
			return src == "image/png" && tgt == "image/webp"
		}})

		got := r.Formats(known)
		want := map[convert.MediaType][]convert.MediaType{
			"image/png": {"image/webp"},
			// image/webp and image/avif have no valid target and must not
			// appear as keys at all.
		}

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Formats() mismatch (-want +got):\n%s", diff)
		}
		if _, ok := got["image/avif"]; ok {
			t.Error(`expected "image/avif" to be absent from the map, not present with an empty slice`)
		}
	})
}

func TestRegistry_StartAll(t *testing.T) {
	t.Run("only Lifecycle transformers are started", func(t *testing.T) {
		plain := &fakeConverter{}
		lc := &fakeLifecycleConverter{}

		r := &convert.Registry{}
		r.Register(plain)
		r.Register(lc)

		if err := r.StartAll(); err != nil {
			t.Fatalf("StartAll() = %v, want nil", err)
		}
		if !lc.started {
			t.Error("expected the Lifecycle transformer to be started")
		}
	})

	t.Run("aggregates errors from every transformer instead of stopping at the first", func(t *testing.T) {
		errA := errors.New("start A failed")
		errB := errors.New("start B failed")
		a := &fakeLifecycleConverter{startErr: errA}
		b := &fakeLifecycleConverter{startErr: errB}

		r := &convert.Registry{}
		r.Register(a)
		r.Register(b)

		err := r.StartAll()
		if !errors.Is(err, errA) {
			t.Errorf("StartAll() error does not wrap %v", errA)
		}
		if !errors.Is(err, errB) {
			t.Errorf("StartAll() error does not wrap %v", errB)
		}
		if !a.started || !b.started {
			t.Error("expected both transformers to have Start() called despite errors")
		}
	})
}

func TestRegistry_StopAll(t *testing.T) {
	t.Run("only Lifecycle transformers are stopped", func(t *testing.T) {
		plain := &fakeConverter{}
		lc := &fakeLifecycleConverter{}

		r := &convert.Registry{}
		r.Register(plain)
		r.Register(lc)

		if err := r.StopAll(); err != nil {
			t.Fatalf("StopAll() = %v, want nil", err)
		}
		if !lc.stopped {
			t.Error("expected the Lifecycle transformer to be stopped")
		}
	})

	t.Run("aggregates errors from every transformer instead of stopping at the first", func(t *testing.T) {
		errA := errors.New("stop A failed")
		errB := errors.New("stop B failed")
		a := &fakeLifecycleConverter{stopErr: errA}
		b := &fakeLifecycleConverter{stopErr: errB}

		r := &convert.Registry{}
		r.Register(a)
		r.Register(b)

		err := r.StopAll()
		if !errors.Is(err, errA) {
			t.Errorf("StopAll() error does not wrap %v", errA)
		}
		if !errors.Is(err, errB) {
			t.Errorf("StopAll() error does not wrap %v", errB)
		}
		if !a.stopped || !b.stopped {
			t.Error("expected both transformers to have Stop() called despite errors")
		}
	})

	t.Run("stops transformers in reverse registration order", func(t *testing.T) {
		var order []string
		first := &fakeLifecycleConverter{name: "first", order: &order}
		second := &fakeLifecycleConverter{name: "second", order: &order}

		r := &convert.Registry{}
		r.Register(first)
		r.Register(second)

		if err := r.StopAll(); err != nil {
			t.Fatalf("StopAll() = %v, want nil", err)
		}

		want := []string{"second", "first"}
		if diff := cmp.Diff(want, order); diff != "" {
			t.Errorf("stop order mismatch (-want +got):\n%s", diff)
		}
	})
}
