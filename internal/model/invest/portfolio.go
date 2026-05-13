package invest

import (
	"context"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"
	"github.com/sevlyar/box"
)

type PortfolioFetcher interface {
	// FetchPortfolio fetches a portfolio by the specified account.
	FetchPortfolio(ctx context.Context, p *Portfolio) error
}

type Portfolio struct {
	Account   AccountRef
	Positions []PortfolioPosition
}

type PortfolioPosition struct {
	Instrument      instrument.Instrument
	Quantity        int64
	Ticker          string
	ClassCode       string
	AveragePrice    instrument.Money
	InstrumentPrice instrument.Money
	ACI             box.Optional[instrument.Money]
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

func (r *PortfolioRegistry) GetPortfolio(ctx context.Context, ref AccountRef) (*Portfolio, error) {
	p := &Portfolio{Account: ref}
	if err := r.portfolios.FetchPortfolio(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}
