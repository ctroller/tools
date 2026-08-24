package convert

import (
	"errors"
	"log/slog"
	"slices"
)

type Registry struct {
	transformers []Converter
}

func (r *Registry) Register(c Converter) {
	r.transformers = append(r.transformers, c)
}

func (r *Registry) Lookup(src, tgt MediaType) (Converter, bool) {
	for _, t := range r.transformers {
		if slices.Contains(t.SupportedTargets(src), tgt) {
			return t, true
		}
	}

	return nil, false
}

func (r *Registry) Formats(src MediaType) []MediaType {
	check := make(map[MediaType]bool)
	for _, t := range r.transformers {
		for _, tgt := range t.SupportedTargets(src) {
			if _, exists := check[tgt]; !exists {
				check[tgt] = true
			}
		}
	}

	formats := make([]MediaType, 0, len(check))
	for format := range check {
		formats = append(formats, format)
	}

	return formats
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

func NewRegistry() *Registry {
	return &Registry{}
}
