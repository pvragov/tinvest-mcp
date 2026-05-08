package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"
	"github.com/sevlyar/box"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type BondService interface {
	GetBondCoupons(
		ctx context.Context, bond instrument.BondRef, params instrument.GetBondCouponsParams,
	) ([]instrument.BondCoupon, error)
	GetBondRedemptions(
		ctx context.Context, bond instrument.BondRef, params instrument.GetBondRedemptionParams,
	) ([]instrument.BondRedemption, error)
	GetBond(ctx context.Context, ref instrument.BondRef) (*instrument.Bond, error)
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
				Coupons: make([]bondCouponView, len(coupons)),
			}

			for i := range coupons {
				reply.Coupons[i] = mapCoupon(&coupons[i])
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
	PayDate        *time.Time        `json:"payDate"`
	Period         *couponPeriodView `json:"period"`
	PeriodDayCount int               `json:"periodDayCount"`
	No             int               `json:"no"`
	OneBondPay     moneyView         `json:"oneBondPay"`
}

func mapCoupon(c *instrument.BondCoupon) bondCouponView {
	ret := bondCouponView{
		No:             c.No,
		PeriodDayCount: int(c.PeriodDays),
		OneBondPay:     moneyView(c.OneBondPay),
	}

	if c.PayDate.IsSome() {
		ret.PayDate = new(c.PayDate.Get())
	}

	if c.Period.IsSome() {
		ret.Period = new(couponPeriodView(c.Period.Get()))
	}

	return ret
}

type couponPeriodView struct {
	Start time.Time `json:"startDate"`
	End   time.Time `json:"endDate"`
}

func NewGetBondRedemptionsTool(service BondService) server.ServerTool {
	const (
		instrumentArgName = "instrument-id"
		fromArgName       = "from"
		toArgName         = "to"
	)

	return server.ServerTool{
		Tool: mcp.NewTool(
			"tbank-get-bond-redemptions",
			mcp.WithDescription("Позволяет получить список погашений для облигации за указанный период"),
			mcp.WithString(instrumentArgName, mcp.Description("Идентификатор инструмента (облигации)"), mcp.Required()),
			mcp.WithString(fromArgName, mcp.Description("Начало периода в формате YYYY-MM-DDTHH:MM:SSZ"), mcp.Required()),
			mcp.WithString(toArgName, mcp.Description("Конец периода в формате YYYY-MM-DDTHH:MM:SSZ"), mcp.Required()),
			mcp.WithOutputSchema[getBondRedemptionsReply](),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			tr, err := parseTimeRangeArgs(req.GetString(fromArgName, ""), req.GetString(toArgName, ""))
			if err != nil {
				return nil, err
			}

			ref := instrument.BondRef{ID: req.GetString(instrumentArgName, "")}

			reds, err := service.GetBondRedemptions(ctx, ref, instrument.GetBondRedemptionParams(tr))
			if err != nil {
				return nil, fmt.Errorf("failed to get bond redemptions: %w", err)
			}

			reply := getBondRedemptionsReply{
				ID:          ref.ID,
				Redemptions: make([]bondRedemptionView, len(reds)),
			}

			for i := range reds {
				reply.Redemptions[i] = mapBondRedemption(reds[i])
			}

			return mcp.NewToolResultJSON(reply)
		},
	}
}

type getBondRedemptionsReply struct {
	ID          string               `json:"instrumentID"`
	Redemptions []bondRedemptionView `json:"redemptions"`
}

type bondRedemptionView struct {
	PayDate    time.Time `json:"payDate"`
	OneBondPay moneyView `json:"oneBondPay"`
}

func mapBondRedemption(r instrument.BondRedemption) bondRedemptionView {
	return bondRedemptionView{
		PayDate:    r.PayDate,
		OneBondPay: moneyView(r.OneBondPay),
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
				MaturityDate:      getOptionalValue(bond.MaturityDate),
				LotSize:           bond.LotSize,
				Nominal:           moneyView(bond.Nominal),
				InitialNominal:    moneyView(bond.InitialNominal),
				ACI:               (*moneyView)(getOptionalValue(bond.ACI)),
				HasAmortization:   bond.HasAmortization,
				HasFloatingCoupon: bond.HasFloatingCoupon,
			})
		},
	}
}

type getBondReply struct {
	ID                string     `json:"instrumentID"`
	Name              string     `json:"name"`
	ISIN              string     `json:"isin"`
	Currency          string     `json:"currency"`
	LotSize           int        `json:"lotSize"`
	Nominal           moneyView  `json:"currentNominalPrice"`
	InitialNominal    moneyView  `json:"initialNominalPrice"`
	ACI               *moneyView `json:"accruedCouponInterest"`
	MaturityDate      *time.Time `json:"maturityDate"`
	HasAmortization   bool       `json:"hasAmortization"`
	HasFloatingCoupon bool       `json:"hasFloatingCoupon"`
}

func getOptionalValue[T any](v box.Optional[T]) *T {
	if v.IsSome() {
		return new(v.Get())
	}

	return nil
}
