package mcp

import (
	"context"
	"fmt"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"
	"github.com/pvragov/tinvest-mcp/internal/model/invest"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type PortfolioService interface {
	GetPortfolio(ctx context.Context, ref invest.Ref) (*invest.Portfolio, error)
}

func NewGetPortfolio(service PortfolioService) server.ServerTool {
	const accountIDArgName = "account-id"

	return server.ServerTool{
		Tool: mcp.NewTool(
			"get-portfolio",
			mcp.WithDescription("Позволяет получить портфель пользователя по номеру счета"),
			mcp.WithString(accountIDArgName, mcp.Description("Идентификатор счета"), mcp.Required()),
			mcp.WithOutputSchema[getUserPortfolioReply](),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := service.GetPortfolio(ctx, invest.Ref{ID: req.GetString(accountIDArgName, "")})
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
	InstrumentID string          `json:"instrumentID"`
	FIGI         string          `json:"FIGI"`
	Quantity     int64           `json:"quantity"`
	Instrument   instrument.Type `json:"instrumentType"`
	Ticker       string          `json:"ticker"`
	ClassCode    string          `json:"classCode"`
}

func mapPortfolio(p *invest.Portfolio) portfolioView {
	view := portfolioView{
		AccountID: p.Account.ID,
		Positions: make([]positionView, len(p.Positions)),
	}

	for i, pos := range p.Positions {
		view.Positions[i] = positionView{
			InstrumentID: pos.ID,
			FIGI:         pos.FIGI,
			Quantity:     pos.Quantity,
			Instrument:   pos.Instrument,
			Ticker:       pos.Ticker,
			ClassCode:    pos.ClassCode,
		}
	}

	return view
}
