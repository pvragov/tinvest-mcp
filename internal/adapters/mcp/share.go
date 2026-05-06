package mcp

import (
	"context"
	"fmt"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ShareService interface {
	GetShare(ctx context.Context, ref instrument.ShareRef) (*instrument.Share, error)
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
				return nil, fmt.Errorf("failed to get bond: %w", err)
			}

			return mcp.NewToolResultJSON(getShareReply{
				ID:       share.ID,
				Name:     share.Name,
				ISIN:     share.ISIN,
				Currency: share.Currency,
			})
		},
	}
}

type getShareReply struct {
	ID       string `json:"instrumentID"`
	Name     string `json:"name"`
	ISIN     string `json:"isin"`
	Currency string `json:"currency"`
}
