package convert

import (
	"errors"
	"log/slog"
	"slices"
)

type Registry struct {
	transformers []Converter
	owners       map[MediaType]map[MediaType]Converter
}

func (r *Registry) Register(c Converter) {
	r.transformers = append(r.transformers, c)
	for src, targets := range c.SupportedFormats() {
		if r.owners[src] == nil {
			r.owners[src] = make(map[MediaType]Converter)
		}
		for _, tgt := range targets {
			if _, claimed := r.owners[src][tgt]; !claimed {
				r.owners[src][tgt] = c // first-registered converter to claim (src,tgt) wins
			} else {
				slog.Warn("Converter already claimed - will not be used for target mapping", "converter", c.Name(), "src", src, "tgt", tgt)
			}
		}
	}
}

func (r *Registry) Lookup(src, tgt MediaType) (Converter, bool) {
	conv, ok := r.owners[src][tgt]
	return conv, ok
}

func (r *Registry) Supports(src MediaType) ([]MediaType, bool) {
	t, ok := r.Formats()[src]
	return t, ok
}

func (r *Registry) Formats() map[MediaType][]MediaType {
	matrix := make(map[MediaType][]MediaType, len(r.owners))
	for src, targets := range r.owners {
		list := make([]MediaType, 0, len(targets))
		for tgt := range targets {
			list = append(list, tgt)
		}

		slices.Sort(list)
		matrix[src] = list
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
	for _, t := range slices.Backward(r.transformers) {

		if lc, ok := t.(Lifecycle); ok {
			slog.Info("Stopping converter lifecycle", "converter", t.Name())
			if err := lc.Stop(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

func (r *Registry) List() []Converter {
	return slices.Clone(r.transformers)
}

func NewRegistry() *Registry {
	return &Registry{
		owners: make(map[MediaType]map[MediaType]Converter),
	}
}
