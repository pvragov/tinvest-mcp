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
	GetBondCoupons(ctx context.Context, bond instrument.BondRef, params instrument.GetBondCouponsParams) ([]instrument.BondCoupon, error)
	GetBond(ctx context.Context, bond *instrument.Bond) error
}

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
			fromTime, err := time.Parse(time.RFC3339, req.GetString(fromArgName, ""))
			if err != nil {
				return nil, fmt.Errorf("invalid 'from' arg: %w", err)
			}

			toTime, err := time.Parse(time.RFC3339, req.GetString(toArgName, ""))
			if err != nil {
				return nil, fmt.Errorf("invalid 'to' arg: %w", err)
			}

			coupons, err := service.GetBondCoupons(ctx, instrument.BondRef{
				ID: req.GetString(instrumentArgName, ""),
			}, instrument.GetBondCouponsParams{
				From: fromTime,
				To:   toTime,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get bond coupons: %w", err)
			}

			reply := getBondCouponsReply{Coupons: make([]bondCouponView, len(coupons))}
			for i, coupon := range coupons {
				reply.Coupons[i] = mapCoupon(&coupon)
			}

			return mcp.NewToolResultJSON(reply)
		},
	}
}

type getBondCouponsReply struct {
	Coupons []bondCouponView `json:"coupons"`
}

type bondCouponView struct {
	CouponDate       time.Time `json:"couponDate"`
	CouponNumber     int       `json:"couponNumber"`
	CouponPeriodDays int32     `json:"couponPeriodDays"`
	OneBondPay       moneyView `json:"oneBondPay"`
}

type moneyView struct {
	Unit      int64  `json:"unit"`
	MinorUnit int32  `json:"minorUnit"`
	Currency  string `json:"currency"`
}

func mapCoupon(c *instrument.BondCoupon) bondCouponView {
	return bondCouponView{
		CouponDate:       c.CouponDate,
		CouponNumber:     c.CouponNumber,
		CouponPeriodDays: c.CouponPeriodDays,
		OneBondPay: moneyView{
			Unit:      c.OneBondPay.Units,
			MinorUnit: c.OneBondPay.MinorUnits,
			Currency:  c.OneBondPay.Currency,
		},
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
			bond := &instrument.Bond{ID: req.GetString(instrumentArgName, "")}
			if err := service.GetBond(ctx, bond); err != nil {
				return nil, fmt.Errorf("failed to get bond: %w", err)
			}

			return mcp.NewToolResultJSON(getBondReply(*bond))
		},
	}
}

type getBondReply struct {
	ID              string `json:"instrumentID"`
	Name            string `json:"name"`
	ISIN            string `json:"isin"`
	Currency        string `json:"currency"`
	HasAmortization bool   `json:"hasAmortization"`
}
