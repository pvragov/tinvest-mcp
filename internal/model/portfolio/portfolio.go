package portfolio

import (
	"context"

	"tinkoff-invest-mcp/internal/model/instrument"
	"tinkoff-invest-mcp/internal/model/user"
)

type Fetcher interface {
	// FetchPortfolio fetches a portfolio by the specified account.
	FetchPortfolio(ctx context.Context, p *Portfolio) error
}

type Portfolio struct {
	Account   user.Ref
	Positions []Position
}

type Position struct {
	ID         string
	FIGI       string
	Quantity   int64
	Instrument instrument.Type
	Ticker     string
	ClassCode  string
}

type Repository interface {
	Fetcher
}

type Registry struct {
	portfolios Repository
}

func NewRegistry(portfolios Repository) *Registry {
	return &Registry{
		portfolios: portfolios,
	}
}

func (r *Registry) GetPortfolio(ctx context.Context, ref user.Ref) (*Portfolio, error) {
	p := &Portfolio{Account: ref}
	if err := r.portfolios.FetchPortfolio(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}
