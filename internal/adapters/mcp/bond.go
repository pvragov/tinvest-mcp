package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type BondService interface {
	GetBondCoupons(
		ctx context.Context, bond instrument.BondRef, params instrument.GetBondCouponsParams,
	) ([]instrument.BondCoupon, error)
	GetBond(ctx context.Context, ref instrument.BondRef) (*instrument.Bond, error)
}

//nolint:dupl
func NewGetBondCouponsTool(service BondService) server.ServerTool {
	const (
		instrumentArgName = "instrument-id"
		fromArgName       = "from"
		toArgName         = "to"
	)

	return server.ServerTool{
		Tool: mcp.NewTool(
			"tbank-get-bond-coupons",
			mcp.WithDescription("Позволяет получить список купонов для облигации за указанный период"),
			mcp.WithString(instrumentArgName, mcp.Description("Идентификатор инструмента (облигации)"), mcp.Required()),
			mcp.WithString(fromArgName, mcp.Description("Начало периода в формате YYYY-MM-DDTHH:MM:SSZ"), mcp.Required()),
			mcp.WithString(toArgName, mcp.Description("Конец периода в формате YYYY-MM-DDTHH:MM:SSZ"), mcp.Required()),
			mcp.WithOutputSchema[getBondCouponsReply](),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tr, err := parseTimeRangeArgs(req.GetString(fromArgName, ""), req.GetString(toArgName, ""))
			if err != nil {
				return nil, err
			}

			ref := instrument.BondRef{ID: req.GetString(instrumentArgName, "")}

			coupons, err := service.GetBondCoupons(ctx, ref, instrument.GetBondCouponsParams(tr))
			if err != nil {
				return nil, fmt.Errorf("failed to get bond coupons: %w", err)
			}

			reply := getBondCouponsReply{
				ID:      ref.ID,
				Coupons: make([]bondCouponView, len(coupons))}
			for i, coupon := range coupons {
				reply.Coupons[i] = mapCoupon(&coupon)
			}

			return mcp.NewToolResultJSON(reply)
		},
	}
}

type timeRange struct {
	From time.Time
	To   time.Time
}

func parseTimeRangeArgs(from, to string) (timeRange, error) {
	fromTime, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return timeRange{}, fmt.Errorf("invalid 'from' arg: %w", err)
	}

	toTime, err := time.Parse(time.RFC3339, to)
	if err != nil {
		return timeRange{}, fmt.Errorf("invalid 'to' arg: %w", err)
	}

	return timeRange{From: fromTime, To: toTime}, nil
}

type getBondCouponsReply struct {
	ID      string           `json:"instrumentID"`
	Coupons []bondCouponView `json:"coupons"`
}

type bondCouponView struct {
	CouponDate       time.Time `json:"couponDate"`
	CouponNumber     int       `json:"couponNumber"`
	CouponPeriodDays int32     `json:"couponPeriodDays"`
	OneBondPay       moneyView `json:"oneBondPay"`
}

func mapCoupon(c *instrument.BondCoupon) bondCouponView {
	return bondCouponView{
		CouponDate:       c.CouponDate,
		CouponNumber:     c.CouponNumber,
		CouponPeriodDays: c.CouponPeriodDays,
		OneBondPay:       moneyView(c.OneBondPay),
	}
}

func NewGetBondTool(service BondService) server.ServerTool {
	const instrumentArgName = "instrument-id"

	return server.ServerTool{
		Tool: mcp.NewTool(
			"tbank-get-bond",
			mcp.WithDescription("Позволяет получить общую информацию по облигации"),
			mcp.WithString(instrumentArgName, mcp.Description("Идентификатор инструмента (облигации)"), mcp.Required()),
			mcp.WithOutputSchema[getBondReply](),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ref := instrument.BondRef{ID: req.GetString(instrumentArgName, "")}

			bond, err := service.GetBond(ctx, ref)
			if err != nil {
				return nil, fmt.Errorf("failed to get bond: %w", err)
			}

			return mcp.NewToolResultJSON(getBondReply{
				ID:                bond.ID,
				Name:              bond.Name,
				ISIN:              bond.ISIN,
				Currency:          bond.Currency,
				HasAmortization:   bond.HasAmortization,
				HasFloatingCoupon: bond.HasFloatingCoupon,
				LotSize:           bond.LotSize,
				Nominal:           moneyView(bond.Nominal),
				InitialNominal:    moneyView(bond.InitialNominal),
			})
		},
	}
}

type getBondReply struct {
	ID                string    `json:"instrumentID"`
	Name              string    `json:"name"`
	ISIN              string    `json:"isin"`
	Currency          string    `json:"currency"`
	LotSize           int       `json:"lotSize"`
	Nominal           moneyView `json:"nominalPrice"`
	InitialNominal    moneyView `json:"initialNominalPrice"`
	HasAmortization   bool      `json:"hasAmortization"`
	HasFloatingCoupon bool      `json:"hasFloatingCoupon"`
}
