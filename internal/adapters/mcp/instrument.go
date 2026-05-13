package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pvragov/tinvest-mcp/internal/model/instrument"
)

type InstrumentService interface {
	SearchInstrument(ctx context.Context, query string) ([]instrument.Info, error)
}

func NewSearchInstrumentTool(service InstrumentService) server.ServerTool {
	searchQueryArgName := "query"

	return server.ServerTool{
		Tool: mcp.NewTool(
			"tbank-instrument-search",
			mcp.WithDescription("Позволяет выполнить поиск инструмента (ценной бумаги) по подстроке"),
			mcp.WithString(
				"query",
				mcp.MinLength(3),
				mcp.Description("Подстрока для поиска (ISIN, наименование бумаги)"),
				mcp.Required(),
			),
			mcp.WithOutputSchema[instrumentSearchReply](),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			infos, err := service.SearchInstrument(ctx, req.GetString(searchQueryArgName, ""))
			if err != nil {
				return nil, fmt.Errorf("failed to search instruments: %w", err)
			}

			reply := instrumentSearchReply{
				Instruments: make([]instrumentInfoView, len(infos)),
			}
			for i := range infos {
				reply.Instruments[i] = instrumentInfoView{
					instrumentView: mapInstrument(&infos[i].Instrument),
					ISIN:           infos[i].ISIN,
					Name:           infos[i].Name,
				}
			}

			return mcp.NewToolResultJSON(reply)
		},
	}
}

type instrumentSearchReply struct {
	Instruments []instrumentInfoView `json:"instruments"`
}

type instrumentInfoView struct {
	instrumentView
	Name string `json:"name"`
	ISIN string `json:"isin"`
}

type instrumentView struct {
	Type      string `json:"instrumentType"`
	ID        string `json:"instrumentID"`
	Ticker    string `json:"ticker"`
	ClassCode string `json:"classCode"`
}

func mapInstrument(i *instrument.Instrument) instrumentView {
	return instrumentView{
		Type:      i.Type.String(),
		ID:        i.ID,
		Ticker:    i.Ticker,
		ClassCode: i.ClassCode,
	}
}
