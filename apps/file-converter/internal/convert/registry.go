package convert

import (
	"errors"
	"log/slog"
)

type Registry struct {
	transformers []Converter
}

func (r *Registry) Register(c Converter) {
	r.transformers = append(r.transformers, c)
}

func (r *Registry) Lookup(src, tgt MediaType) (Converter, bool) {
	for _, t := range r.transformers {
		if t.CanHandle(src, tgt) {
			return t, true
		}
	}
	return nil, false
}

func (r *Registry) Formats(known []MediaType) map[MediaType][]MediaType {
	matrix := make(map[MediaType][]MediaType)
	for _, src := range known {
		for _, tgt := range known {
			if src != tgt {
				if _, ok := r.Lookup(src, tgt); ok {
					matrix[src] = append(matrix[src], tgt)
				}
			}
		}
	}
	return matrix
}

func (r *Registry) StartAll() error {
	for _, t := range r.transformers {
		if lc, ok := t.(Lifecycle); ok {
			slog.Info("Starting up converter lifecycle", "converter", lc)
			if err := lc.Start(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Registry) StopAll() error {
	var errs []error
	for i := len(r.transformers) - 1; i >= 0; i-- {
		if lc, ok := r.transformers[i].(Lifecycle); ok {
			slog.Info("Stopping converter lifecycle", "converter", lc)
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
