package convert

import (
	"errors"
	"log/slog"
	"slices"
)

type Registry struct {
	transformers []Converter
	owners       map[MediaType]map[MediaType]Converter
	fmMatrix     map[MediaType][]MediaType
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
			}
		}
	}

	r.invalidateFormats()
}

func (r *Registry) Lookup(src, tgt MediaType) (Converter, bool) {
	conv, ok := r.owners[src][tgt]
	return conv, ok
}

func (r *Registry) Formats() map[MediaType][]MediaType {
	c := make(map[MediaType][]MediaType, len(r.fmMatrix))
	for src, targets := range r.fmMatrix {
		c[src] = slices.Clone(targets)
	}
	return c
}

func (r *Registry) invalidateFormats() {
	matrix := make(map[MediaType][]MediaType, len(r.owners))
	for src, targets := range r.owners {
		for tgt := range targets {
			matrix[src] = append(matrix[src], tgt)
		}

		slices.Sort(matrix[src])
	}

	r.fmMatrix = matrix
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
	return r.transformers
}

func NewRegistry() *Registry {
	return &Registry{
		owners: make(map[MediaType]map[MediaType]Converter),
	}
}
