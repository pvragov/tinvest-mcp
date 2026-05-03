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
			mcp.WithOutputSchema[GetUserPortfolioReply](),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			p, err := service.GetPortfolio(ctx, invest.Ref{ID: req.GetString(accountIDArgName, "")})
			if err != nil {
				return nil, fmt.Errorf("failed to get portfolio: %w", err)
			}

			return mcp.NewToolResultJSON(GetUserPortfolioReply{
				Portfolio: mapPortfolio(p),
			})
		},
	}
}

type GetUserPortfolioReply struct {
	Portfolio PortfolioView `json:"portfolio"`
}

type PortfolioView struct {
	AccountID string         `json:"accountID"`
	Positions []PositionView `json:"positions"`
}

type PositionView struct {
	InstrumentID string          `json:"instrumentID"`
	FIGI         string          `json:"FIGI"`
	Quantity     int64           `json:"quantity"`
	Instrument   instrument.Type `json:"instrumentType"`
	Ticker       string          `json:"ticker"`
	ClassCode    string          `json:"classCode"`
}

func mapPortfolio(p *invest.Portfolio) PortfolioView {
	view := PortfolioView{
		AccountID: p.Account.ID,
		Positions: make([]PositionView, len(p.Positions)),
	}

	for i, pos := range p.Positions {
		view.Positions[i] = PositionView{
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
