package convert

import (
	"errors"
	"log/slog"
)

type Registry struct {
	transformers []Converter
	owners       map[MediaType]map[MediaType]Converter
}

func (r *Registry) Register(c Converter) {
	r.transformers = append(r.transformers, c)
}

func (r *Registry) Lookup(src, tgt MediaType) (Converter, bool) {
	conv, ok := r.owners[src][tgt]
	return conv, ok
}

// Formats is derived from owners, not from re-walking transformers directly.
// Doing the latter previously aliased a converter's own cached SupportedFormats
// slice and appended into it across converters, corrupting that converter's
// state; deriving from owners also guarantees this never advertises a pair
// that Lookup wouldn't actually honor.
func (r *Registry) Formats() map[MediaType][]MediaType {
	matrix := make(map[MediaType][]MediaType, len(r.owners))
	for src, targets := range r.owners {
		for tgt := range targets {
			matrix[src] = append(matrix[src], tgt)
		}
	}
	return matrix
}

func (r *Registry) StartAll() error {
	var errs []error
	for _, t := range r.transformers {
		if lc, ok := t.(Lifecycle); ok {
			slog.Info("Starting up converter lifecycle", "converter", t.Name())
			if err := lc.Start(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (r *Registry) StopAll() error {
	var errs []error
	for i := len(r.transformers) - 1; i >= 0; i-- {
		t := r.transformers[i]
		if lc, ok := t.(Lifecycle); ok {
			slog.Info("Stopping converter lifecycle", "converter", t.Name())
			if err := lc.Stop(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

func (r *Registry) GetAll() []Converter {
	return r.transformers
}

// Build snapshots the currently registered transformers into owners, powering
// Lookup and Formats. It must run after the last Register call and before the
// registry is used to serve requests: Lookup/Formats silently read an empty
// owners map if called first, and a transformer registered after Build has
// already run stays invisible until Build runs again.
func (r *Registry) Build() {
	r.owners = make(map[MediaType]map[MediaType]Converter)
	for _, t := range r.transformers {
		for src, targets := range t.SupportedFormats() {
			if r.owners[src] == nil {
				r.owners[src] = make(map[MediaType]Converter)
			}
			for _, tgt := range targets {
				if _, claimed := r.owners[src][tgt]; !claimed {
					r.owners[src][tgt] = t // first-registered converter to claim (src,tgt) wins
				}
			}
		}
	}
}

func NewRegistry() *Registry {
	return &Registry{}
}
