package mcp

import (
	"context"
	"fmt"
	"time"

	"tinkoff-invest-mcp/internal/model/instrument"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type BondService interface {
	GetBondCoupons(ctx context.Context, bond instrument.BondRef, params instrument.GetBondCouponsParams) ([]instrument.BondCoupon, error)
}

func NewGetBondCouponsTool(service BondService) server.ServerTool {
	const (
		instrumentArgName = "instrument-id"
		fromArgName       = "from"
		toArgName         = "to"
	)

	return server.ServerTool{
		Tool: mcp.NewTool(
			"get-bond-coupons",
			mcp.WithDescription("Позволяет получить список купонов для облигации за указанный период"),
			mcp.WithString(instrumentArgName, mcp.Description("Идентификатор инструмента (облигации)"), mcp.Required()),
			mcp.WithString(fromArgName, mcp.Description("Начало периода"), mcp.Required()),
			mcp.WithString(toArgName, mcp.Description("Конец периода"), mcp.Required()),
			mcp.WithOutputSchema[GetBondCouponsReply](),
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
				return nil, fmt.Errorf("failed to get portfolio: %w", err)
			}

			reply := GetBondCouponsReply{Coupons: make([]BondCouponView, len(coupons))}
			for i, coupon := range coupons {
				reply.Coupons[i] = mapCoupon(&coupon)
			}

			return mcp.NewToolResultJSON(reply)
		},
	}
}

type BondCouponView struct {
	FIGI             string    `json:"figi"`
	CouponDate       time.Time `json:"couponDate"`
	CouponStartDate  time.Time `json:"couponStartDate"`
	CouponEndDate    time.Time `json:"couponEndDate"`
	CouponNumber     int       `json:"couponNumber"`
	CouponPeriodDays int32     `json:"couponPeriodDays"`
	OneBondPay       string    `json:"oneBondPay"`
}

func mapCoupon(c *instrument.BondCoupon) BondCouponView {
	return BondCouponView{
		FIGI:             c.FIGI,
		CouponDate:       c.CouponDate,
		CouponStartDate:  c.CouponStartDate,
		CouponEndDate:    c.CouponEndDate,
		CouponNumber:     c.CouponNumber,
		CouponPeriodDays: c.CouponPeriodDays,
		OneBondPay:       c.OneBondPay.String(),
	}
}

type GetBondCouponsReply struct {
	Coupons []BondCouponView `json:"coupons"`
}
