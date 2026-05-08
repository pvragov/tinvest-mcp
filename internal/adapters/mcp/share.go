package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ShareService interface {
	GetShare(ctx context.Context, ref instrument.ShareRef) (*instrument.Share, error)
	GetShareDividends(
		ctx context.Context, share instrument.ShareRef, params instrument.GetShareDividendsParams,
	) ([]instrument.Dividend, error)
}

func NewGetShareTool(service ShareService) server.ServerTool {
	const instrumentArgName = "instrument-id"

	return server.ServerTool{
		Tool: mcp.NewTool(
			"tbank-get-share",
			mcp.WithDescription("Позволяет получить общую информацию по акции"),
			mcp.WithString(instrumentArgName, mcp.Description("Идентификатор инструмента (акции)"), mcp.Required()),
			mcp.WithOutputSchema[getShareReply](),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ref := instrument.ShareRef{ID: req.GetString(instrumentArgName, "")}

			share, err := service.GetShare(ctx, ref)
			if err != nil {
				return nil, fmt.Errorf("failed to get share: %w", err)
			}

			return mcp.NewToolResultJSON(getShareReply{
				ID:       share.ID,
				Name:     share.Name,
				ISIN:     share.ISIN,
				Currency: share.Currency,
				LotSize:  share.LotSize,
			})
		},
	}
}

type getShareReply struct {
	ID       string `json:"instrumentID"`
	Name     string `json:"name"`
	ISIN     string `json:"isin"`
	Currency string `json:"currency"`
	LotSize  int    `json:"lotSize"`
}

func NewGetShareDividendsTool(service ShareService) server.ServerTool {
	const (
		instrumentArgName = "instrument-id"
		fromArgName       = "from"
		toArgName         = "to"
	)

	return server.ServerTool{
		Tool: mcp.NewTool(
			"tbank-get-share-dividends",
			mcp.WithDescription("Позволяет получить список дивидендов для акции за указанный период"),
			mcp.WithString(instrumentArgName, mcp.Description("Идентификатор инструмента (акции)"), mcp.Required()),
			mcp.WithString(fromArgName, mcp.Description("Начало периода в формате YYYY-MM-DDTHH:MM:SSZ"), mcp.Required()),
			mcp.WithString(toArgName, mcp.Description("Конец периода в формате YYYY-MM-DDTHH:MM:SSZ"), mcp.Required()),
			mcp.WithOutputSchema[getShareDividendsReply](),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tr, err := parseTimeRangeArgs(req.GetString(fromArgName, ""), req.GetString(toArgName, ""))
			if err != nil {
				return nil, err
			}

			ref := instrument.ShareRef{ID: req.GetString(instrumentArgName, "")}

			dividends, err := service.GetShareDividends(ctx, ref, instrument.GetShareDividendsParams(tr))
			if err != nil {
				return nil, fmt.Errorf("failed to get share dividends: %w", err)
			}

			reply := getShareDividendsReply{
				ID:        ref.ID,
				Dividends: make([]shareDividendsView, len(dividends)),
			}

			for i, d := range dividends {
				reply.Dividends[i] = mapShareDividend(&d)
			}

			return mcp.NewToolResultJSON(reply)
		},
	}
}

type getShareDividendsReply struct {
	ID        string               `json:"instrumentID"`
	Dividends []shareDividendsView `json:"dividends"`
}

type shareDividendsView struct {
	Value        moneyView `json:"valuePerShare"`
	PaymentDate  time.Time `json:"paymentDate"`
	DeclaredDate time.Time `json:"declaredDate"`
	LastBuyDate  time.Time `json:"LastBuyDate"`
	YieldValue   float64   `json:"yieldValuePercent"`
}

func mapShareDividend(d *instrument.Dividend) shareDividendsView {
	return shareDividendsView{
		Value:        moneyView(d.Value),
		PaymentDate:  d.PaymentDate,
		DeclaredDate: d.DeclaredDate,
		LastBuyDate:  d.LastBuyDate,
		YieldValue:   d.YieldValue,
	}
}
