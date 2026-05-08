package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewGetShareTool(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		share := &instrument.Share{
			ID:       "share-1",
			Name:     "Test Share",
			ISIN:     "US0378331005",
			Currency: "usd",
			LotSize:  10,
		}

		service := &MockShareService{}
		service.On("GetShare", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			ref := args.Get(1).(instrument.ShareRef)
			require.Equal(t, "share-1", ref.ID)
		}).Return(share, nil)

		tool := NewGetShareTool(service)

		res, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-share"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"instrument-id": "share-1"},
			},
		})

		require.NoError(t, err)
		require.Equal(t, getShareReply{
			ID:       "share-1",
			Name:     "Test Share",
			ISIN:     "US0378331005",
			Currency: "usd",
			LotSize:  10,
		}, res.StructuredContent)
	})
}

func TestNewGetShareDividendsTool(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		dividends := []instrument.Dividend{{
			Value:        instrument.Money{Units: 1, MinorUnits: 50, Currency: "usd"},
			PaymentDate:  time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			DeclaredDate: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
			LastBuyDate:  time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC),
			YieldValue:   3.5,
		}}

		service := &MockShareService{}
		service.On("GetShareDividends", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			ref := args.Get(1).(instrument.ShareRef)
			params := args.Get(2).(instrument.GetShareDividendsParams)
			require.Equal(t, "share-1", ref.ID)
			require.Equal(t, from, params.From)
			require.Equal(t, to, params.To)
		}).Return(dividends, nil)

		tool := NewGetShareDividendsTool(service)

		res, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-share-dividends"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"instrument-id": "share-1",
					"from":          from.Format(time.RFC3339),
					"to":            to.Format(time.RFC3339),
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, getShareDividendsReply{
			ID: "share-1",
			Dividends: []shareDividendsView{{
				Value:        moneyView{Units: 1, MinorUnits: 50, Currency: "usd"},
				PaymentDate:  time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
				DeclaredDate: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
				LastBuyDate:  time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC),
				YieldValue:   3.5,
			}},
		}, res.StructuredContent)
	})

	t.Run("invalid from", func(t *testing.T) {
		tool := NewGetShareDividendsTool(NewMockShareServiceStub())

		_, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-share-dividends"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"instrument-id": "share-1",
					"from":          "not-a-time",
					"to":            to.Format(time.RFC3339),
				},
			},
		})

		require.Error(t, err)
	})

	t.Run("invalid to", func(t *testing.T) {
		tool := NewGetShareDividendsTool(NewMockShareServiceStub())

		_, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-share-dividends"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"instrument-id": "share-1",
					"from":          from.Format(time.RFC3339),
					"to":            "not-a-time",
				},
			},
		})

		require.Error(t, err)
	})
}

type MockShareService struct {
	mock.Mock
}

func NewMockShareServiceStub() *MockShareService {
	s := &MockShareService{}
	s.On("GetShare", mock.Anything, mock.Anything).Return(&instrument.Share{}, nil)
	s.On("GetShareDividends", mock.Anything, mock.Anything, mock.Anything).Return([]instrument.Dividend{}, nil)
	return s
}

func (m *MockShareService) GetShare(ctx context.Context, ref instrument.ShareRef) (*instrument.Share, error) {
	args := m.Called(ctx, ref)
	return args.Get(0).(*instrument.Share), args.Error(1)
}

func (m *MockShareService) GetShareDividends(ctx context.Context, share instrument.ShareRef, params instrument.GetShareDividendsParams) ([]instrument.Dividend, error) {
	args := m.Called(ctx, share, params)
	return args.Get(0).([]instrument.Dividend), args.Error(1)
}
