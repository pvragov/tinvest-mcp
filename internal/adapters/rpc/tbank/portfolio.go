package tbank

import (
	"context"
	"fmt"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"
	"github.com/pvragov/tinvest-mcp/internal/model/invest"
	"github.com/sevlyar/box"

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

func (a *PortfolioAdapter) FetchPortfolio(_ context.Context, p *invest.Portfolio) error {
	resp, err := a.client.GetPortfolio(p.Account.ID, proto.PortfolioRequest_RUB)
	if err != nil {
		return fmt.Errorf("failed to exec get portfolio rpc: %w", err)
	}

	p.Positions = make([]invest.PortfolioPosition, len(resp.GetPositions()))
	for i, pos := range resp.GetPositions() {
		p.Positions[i] = mapProtoPortfolioPosition(pos)
	}

	return nil
}

func mapProtoPortfolioPosition(p *proto.PortfolioPosition) invest.PortfolioPosition {
	ret := invest.PortfolioPosition{
		Quantity: p.Quantity.Units,
		Instrument: instrument.Instrument{
			Type:      mapInstrumentType[p.InstrumentType],
			ID:        p.GetInstrumentUid(),
			Ticker:    p.GetTicker(),
			ClassCode: p.GetClassCode(),
		},
		Ticker:          p.Ticker,
		ClassCode:       p.ClassCode,
		InstrumentPrice: mapProtoMoney(p.GetCurrentPrice()),
		AveragePrice:    mapProtoMoney(p.GetAveragePositionPrice()),
	}
	if p.CurrentNkd != nil {
		ret.ACI = box.Some(mapProtoMoney(p.CurrentNkd))
	}

	if p.DailyYield != nil {
		ret.DailyYield = box.Some(mapProtoMoney(p.DailyYield))
	}

	if p.ExpectedYield != nil {
		ret.ExpectedYield = box.Some(instrument.Money{
			Units:      p.ExpectedYield.Units,
			MinorUnits: p.ExpectedYield.Nano / 10_000_000,
			Currency:   ret.InstrumentPrice.Currency, // берем валюту из соседнего поля
		})
	}

	return ret
}

var (
	mapInstrumentType = map[string]instrument.Type{
		"bond":     instrument.TypeBond,
		"currency": instrument.TypeCurrency,
		"share":    instrument.TypeShare,
		"etf":      instrument.TypeETF,
	}
)
