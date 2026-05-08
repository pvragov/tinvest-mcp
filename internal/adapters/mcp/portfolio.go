package mcp

import (
	"context"
	"fmt"

	"github.com/pvragov/tinvest-mcp/internal/model/invest"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type PortfolioService interface {
	GetPortfolio(ctx context.Context, ref invest.AccountRef) (*invest.Portfolio, error)
}

func NewGetPortfolio(service PortfolioService) server.ServerTool {
	const accountIDArgName = "account-id"

	return server.ServerTool{
		Tool: mcp.NewTool(
			"tbank-get-portfolio",
			mcp.WithDescription("Позволяет получить портфель пользователя по номеру счета"),
			mcp.WithString(accountIDArgName, mcp.Description("Идентификатор счета"), mcp.Required()),
			mcp.WithOutputSchema[getUserPortfolioReply](),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := service.GetPortfolio(ctx, invest.AccountRef{ID: req.GetString(accountIDArgName, "")})
			if err != nil {
				return nil, fmt.Errorf("failed to get portfolio: %w", err)
			}

			return mcp.NewToolResultJSON(getUserPortfolioReply{
				Portfolio: mapPortfolio(p),
			})
		},
	}
}

type getUserPortfolioReply struct {
	Portfolio portfolioView `json:"portfolio"`
}

type portfolioView struct {
	AccountID string         `json:"accountID"`
	Positions []positionView `json:"positions"`
}

type positionView struct {
	InstrumentID    string     `json:"instrumentID"`
	FIGI            string     `json:"FIGI"`
	Quantity        int64      `json:"quantity"`
	Instrument      string     `json:"instrumentType"`
	Ticker          string     `json:"ticker"`
	ClassCode       string     `json:"classCode"`
	AveragePrice    moneyView  `json:"weightedAveragePrice"`
	InstrumentPrice moneyView  `json:"instrumentCurrentPrice"`
	ACI             *moneyView `json:"accruedCouponInterest"`
}

func mapPortfolio(p *invest.Portfolio) portfolioView {
	view := portfolioView{
		AccountID: p.Account.ID,
		Positions: make([]positionView, len(p.Positions)),
	}

	for i := range p.Positions {
		pos := p.Positions[i]
		view.Positions[i] = positionView{
			InstrumentID:    pos.ID,
			FIGI:            pos.FIGI,
			Quantity:        pos.Quantity,
			Instrument:      pos.Instrument.String(),
			Ticker:          pos.Ticker,
			ClassCode:       pos.ClassCode,
			AveragePrice:    moneyView(pos.AveragePrice),
			InstrumentPrice: moneyView(pos.InstrumentPrice),
		}

		if pos.ACI != nil {
			view.Positions[i].ACI = (*moneyView)(pos.ACI)
		}
	}

	return view
}

type moneyView struct {
	Units      int64  `json:"units"`
	MinorUnits int32  `json:"minorUnits"`
	Currency   string `json:"currency"`
}
