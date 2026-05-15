package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pvragov/tinvest-mcp/internal/model/invest"
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
	instrumentView
	Quantity        int64      `json:"quantity"`
	AveragePrice    moneyView  `json:"weightedAveragePrice"`
	InstrumentPrice moneyView  `json:"instrumentCurrentPrice"`
	ACI             *moneyView `json:"accruedCouponInterest"`
	DailyYield      *moneyView `json:"dailyYield"`
	ExpectedYield   *moneyView `json:"expectedYield"`
}

func mapPortfolio(p *invest.Portfolio) portfolioView {
	view := portfolioView{
		AccountID: p.Account.ID,
		Positions: make([]positionView, len(p.Positions)),
	}

	for i := range p.Positions {
		pos := p.Positions[i]
		view.Positions[i] = positionView{
			instrumentView:  mapInstrument(&pos.Instrument),
			Quantity:        pos.Quantity,
			AveragePrice:    moneyView(pos.AveragePrice),
			InstrumentPrice: moneyView(pos.InstrumentPrice),
		}

		if pos.ACI.IsSome() {
			view.Positions[i].ACI = new(moneyView(pos.ACI.Get()))
		}

		if pos.DailyYield.IsSome() {
			view.Positions[i].DailyYield = new(moneyView(pos.DailyYield.Get()))
		}

		if pos.ExpectedYield.IsSome() {
			view.Positions[i].ExpectedYield = new(moneyView(pos.ExpectedYield.Get()))
		}

	}

	return view
}

type moneyView struct {
	Units      int64  `json:"units"`
	MinorUnits int32  `json:"minorUnits"`
	Currency   string `json:"currency"`
}
