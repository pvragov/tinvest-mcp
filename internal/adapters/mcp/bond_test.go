package mcp

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"
	"github.com/sevlyar/box"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewGetBondTool(t *testing.T) {
	maturityDate := time.Now().UTC()
	aci := newMoneyValue()

	t.Run("success", func(t *testing.T) {
		bond := &instrument.Bond{
			ID:                "bond-1",
			Name:              "Test Bond",
			ISIN:              "RU000A0JWSQ7",
			Currency:          "rub",
			HasAmortization:   true,
			HasFloatingCoupon: true,
			LotSize:           10,
			Nominal:           newMoneyValue(),
			InitialNominal:    newMoneyValue(),
			MaturityDate:      box.Some(maturityDate),
			ACI:               box.Some(aci),
		}

		service := &MockBondService{}
		service.On("GetBond", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			ref := args.Get(1).(instrument.BondRef)
			require.Equal(t, "bond-1", ref.ID)
		}).Return(bond, nil)

		tool := NewGetBondTool(service)

		res, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-bond"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"instrument-id": "bond-1"},
			},
		})

		require.NoError(t, err)
		require.Equal(t, getBondReply{
			ID:                "bond-1",
			Name:              "Test Bond",
			ISIN:              "RU000A0JWSQ7",
			Currency:          "rub",
			HasAmortization:   true,
			HasFloatingCoupon: true,
			LotSize:           10,
			Nominal:           moneyView(bond.Nominal),
			InitialNominal:    moneyView(bond.InitialNominal),
			MaturityDate:      &maturityDate,
			ACI:               (*moneyView)(&aci),
		}, res.StructuredContent)
	})
}

func TestNewGetBondCouponsTool(t *testing.T) {
	from := time.Now().Add(-1 * time.Hour).Truncate(time.Second).UTC()
	to := time.Now().Truncate(time.Second).UTC()

	payDate := time.Now().Add(1 * time.Hour).Truncate(time.Second).UTC()
	period := instrument.CuponPeriod{Start: from, End: to}

	t.Run("success", func(t *testing.T) {
		bondPay := newMoneyValue()
		coupons := []instrument.BondCoupon{{
			PayDate:    box.Some(payDate),
			Period:     box.Some(period),
			No:         1,
			PeriodDays: 182,
			OneBondPay: bondPay,
		}}

		service := &MockBondService{}
		service.On("GetBondCoupons", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			ref := args.Get(1).(instrument.BondRef)
			params := args.Get(2).(instrument.GetBondCouponsParams)
			require.Equal(t, "bond-1", ref.ID)
			require.Equal(t, from, params.From)
			require.Equal(t, to, params.To)
		}).Return(coupons, nil)

		tool := NewGetBondCouponsTool(service)

		res, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-bond-coupons"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"instrument-id": "bond-1",
					"from":          from.Format(time.RFC3339),
					"to":            to.Format(time.RFC3339),
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, getBondCouponsReply{
			ID: "bond-1",
			Coupons: []bondCouponView{{
				PayDate:        &payDate,
				Period:         new(couponPeriodView(period)),
				No:             1,
				PeriodDayCount: 182,
				OneBondPay:     moneyView(bondPay),
			}},
		}, res.StructuredContent)
	})

	t.Run("invalid from", func(t *testing.T) {
		tool := NewGetBondCouponsTool(NewMockBondServiceStub())

		_, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-bond-coupons"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"instrument-id": "bond-1",
					"from":          "not-a-time",
					"to":            to.Format(time.RFC3339),
				},
			},
		})

		require.Error(t, err)
	})

	t.Run("invalid to", func(t *testing.T) {
		tool := NewGetBondCouponsTool(NewMockBondServiceStub())

		_, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-bond-coupons"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"instrument-id": "bond-1",
					"from":          from.Format(time.RFC3339),
					"to":            "not-a-time",
				},
			},
		})

		require.Error(t, err)
	})
}

func TestNewGetBondRedemptionsTool(t *testing.T) {
	from := time.Now().Add(-1 * time.Hour).Truncate(time.Second).UTC()
	to := time.Now().Truncate(time.Second).UTC()

	payDate := time.Now().Add(1 * time.Hour).Truncate(time.Second).UTC()

	t.Run("success", func(t *testing.T) {
		oneBondPay := newMoneyValue()
		redemptions := []instrument.BondRedemption{{
			PayDate:    payDate,
			OneBondPay: oneBondPay,
		}}

		service := &MockBondService{}
		service.On("GetBondRedemptions", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			ref := args.Get(1).(instrument.BondRef)
			params := args.Get(2).(instrument.GetBondRedemptionParams)
			require.Equal(t, "bond-1", ref.ID)
			require.Equal(t, from, params.From)
			require.Equal(t, to, params.To)
		}).Return(redemptions, nil)

		tool := NewGetBondRedemptionsTool(service)

		res, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-bond-redemptions"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"instrument-id": "bond-1",
					"from":          from.Format(time.RFC3339),
					"to":            to.Format(time.RFC3339),
				},
			},
		})

		require.NoError(t, err)
		require.Equal(t, getBondRedemptionsReply{
			ID: "bond-1",
			Redemptions: []bondRedemptionView{{
				PayDate:    payDate,
				OneBondPay: moneyView(oneBondPay),
			}},
		}, res.StructuredContent)
	})

	t.Run("invalid from", func(t *testing.T) {
		tool := NewGetBondRedemptionsTool(NewMockBondServiceStub())

		_, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-bond-redemptions"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"instrument-id": "bond-1",
					"from":          "not-a-time",
					"to":            to.Format(time.RFC3339),
				},
			},
		})

		require.Error(t, err)
	})

	t.Run("invalid to", func(t *testing.T) {
		tool := NewGetBondRedemptionsTool(NewMockBondServiceStub())

		_, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-bond-redemptions"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{
					"instrument-id": "bond-1",
					"from":          from.Format(time.RFC3339),
					"to":            "not-a-time",
				},
			},
		})

		require.Error(t, err)
	})
}

type MockBondService struct {
	mock.Mock
}

func NewMockBondServiceStub() *MockBondService {
	s := &MockBondService{}
	s.On("GetBond", mock.Anything, mock.Anything).Return(&instrument.Bond{}, nil)
	s.On("GetBondCoupons", mock.Anything, mock.Anything, mock.Anything).Return([]instrument.BondCoupon{}, nil)
	s.On("GetBondRedemptions", mock.Anything, mock.Anything, mock.Anything).Return([]instrument.BondRedemption{}, nil)
	return s
}

func (m *MockBondService) GetBond(ctx context.Context, ref instrument.BondRef) (*instrument.Bond, error) {
	args := m.Called(ctx, ref)
	return args.Get(0).(*instrument.Bond), args.Error(1)
}

func (m *MockBondService) GetBondCoupons(ctx context.Context, bond instrument.BondRef, params instrument.GetBondCouponsParams) ([]instrument.BondCoupon, error) {
	args := m.Called(ctx, bond, params)
	return args.Get(0).([]instrument.BondCoupon), args.Error(1)
}

func (m *MockBondService) GetBondRedemptions(ctx context.Context, bond instrument.BondRef, params instrument.GetBondRedemptionParams) ([]instrument.BondRedemption, error) {
	args := m.Called(ctx, bond, params)
	return args.Get(0).([]instrument.BondRedemption), args.Error(1)
}

var moneyValueCounter int32

func newMoneyValue() instrument.Money {
	var (
		units      = atomic.AddInt32(&moneyValueCounter, 1)
		minorUnits = atomic.AddInt32(&moneyValueCounter, 1)
	)
	return instrument.Money{
		Units:      int64(units),
		MinorUnits: minorUnits,
		Currency:   "USD",
	}
}
