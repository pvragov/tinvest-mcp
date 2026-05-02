package tbank

import (
	"context"
	"fmt"
	"tinkoff-invest-mcp/internal/model/instrument"

	"tinkoff-invest-mcp/internal/model/portfolio"

	"opensource.tbank.ru/invest/invest-go/investgo"
	proto "opensource.tbank.ru/invest/invest-go/proto"
)

type PortfolioAdapter struct {
	client *investgo.OperationsServiceClient
}

func NewPortfolioAdapter(client *investgo.OperationsServiceClient) *PortfolioAdapter {
	return &PortfolioAdapter{
		client: client,
	}
}

func (a *PortfolioAdapter) FetchPortfolio(ctx context.Context, p *portfolio.Portfolio) error {
	resp, err := a.client.GetPortfolio(p.Account.ID, proto.PortfolioRequest_RUB)
	if err != nil {
		return fmt.Errorf("failed to exec get portfolio rpc: %w", err)
	}

	p.Positions = make([]portfolio.Position, len(resp.GetPositions()))
	for i, pos := range resp.GetPositions() {
		p.Positions[i] = mapProtoPortfolioPosition(pos)
	}

	return nil
}

func mapProtoPortfolioPosition(p *proto.PortfolioPosition) portfolio.Position {
	return portfolio.Position{
		ID:         p.InstrumentUid,
		FIGI:       p.Figi,
		Quantity:   p.Quantity.Units,
		Instrument: mapInstrumentType[p.InstrumentType],
		Ticker:     p.Ticker,
		ClassCode:  p.ClassCode,
	}
}

var (
	mapInstrumentType = map[string]instrument.Type{
		"bond":     instrument.TypeBond,
		"currency": instrument.TypeCurrency,
		"share":    instrument.TypeShare,
		"etf":      instrument.TypeETF,
	}
)
