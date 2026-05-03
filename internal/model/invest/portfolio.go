package invest

import (
	"context"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"
)

type PortfolioFetcher interface {
	// FetchPortfolio fetches a portfolio by the specified account.
	FetchPortfolio(ctx context.Context, p *Portfolio) error
}

type Portfolio struct {
	Account   Ref
	Positions []PortfolioPosition
}

type PortfolioPosition struct {
	ID         string
	FIGI       string
	Quantity   int64
	Instrument instrument.Type
	Ticker     string
	ClassCode  string
}

type PortfolioRepository interface {
	PortfolioFetcher
}

type PortfolioRegistry struct {
	portfolios PortfolioRepository
}

func NewPortfolioRegistry(portfolios PortfolioRepository) *PortfolioRegistry {
	return &PortfolioRegistry{
		portfolios: portfolios,
	}
}

func (r *PortfolioRegistry) GetPortfolio(ctx context.Context, ref Ref) (*Portfolio, error) {
	p := &Portfolio{Account: ref}
	if err := r.portfolios.FetchPortfolio(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}
