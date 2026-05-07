package mcp

import (
	"context"
	"testing"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"
	"github.com/pvragov/tinvest-mcp/internal/model/invest"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewGetPortfolio(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		portfolio := &invest.Portfolio{
			Account: invest.AccountRef{ID: "acc-1"},
			Positions: []invest.PortfolioPosition{
				{
					ID:         "instr-1",
					FIGI:       "BBG000B9XRY4",
					Quantity:   10,
					Instrument: instrument.TypeShare,
					Ticker:     "AAPL",
					ClassCode:  "SPBXM",
				},
			},
		}

		service := &MockPortfolioService{}
		service.On("GetPortfolio", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			ref := args.Get(1).(invest.AccountRef)
			require.Equal(t, "acc-1", ref.ID)
		}).Return(portfolio, nil)

		tool := NewGetPortfolio(service)

		res, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-portfolio"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"account-id": "acc-1"},
			},
		})

		require.NoError(t, err)
		expected := getUserPortfolioReply{
			Portfolio: portfolioView{
				AccountID: "acc-1",
				Positions: []positionView{{
					InstrumentID: "instr-1",
					FIGI:         "BBG000B9XRY4",
					Quantity:     10,
					Instrument:   instrument.TypeShare.String(),
					Ticker:       "AAPL",
					ClassCode:    "SPBXM",
				}},
			},
		}
		require.Equal(t, expected, res.StructuredContent)
	})
}

type MockPortfolioService struct {
	mock.Mock
}

func (m *MockPortfolioService) GetPortfolio(ctx context.Context, ref invest.AccountRef) (*invest.Portfolio, error) {
	args := m.Called(ctx, ref)
	return args.Get(0).(*invest.Portfolio), args.Error(1)
}
