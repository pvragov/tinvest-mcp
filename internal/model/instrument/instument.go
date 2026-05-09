package instrument

import (
	"context"
	"fmt"
)

type Type int

//go:generate go run github.com/dmarkham/enumer -type=Type -text -json -yaml -transform=lower -trimprefix=Type -output=type_enum.go
const (
	TypeUnknown Type = iota
	TypeShare
	TypeBond
	TypeCurrency
	TypeETF
)

type Info struct {
	Type Type
	ID   string
	ISIN string
	Name string
}

type Searcher interface {
	SearchInstrument(ctx context.Context, query string) ([]Info, error)
}

type Repository interface {
	Searcher
}

type Registry struct {
	repo Searcher
}

func NewRegistry(repo Searcher) *Registry {
	return &Registry{
		repo: repo,
	}
}

func (r *Registry) SearchInstrument(ctx context.Context, query string) ([]Info, error) {
	const queryMinLen = 3
	if query == "" || len(query) < queryMinLen {
		return nil, fmt.Errorf("invalid query")
	}

	return r.repo.SearchInstrument(ctx, query)
}
