package tbank

import (
	"context"
	"tinkoff-invest-mcp/internal/model/instrument"

	"opensource.tbank.ru/invest/invest-go/investgo"
	proto "opensource.tbank.ru/invest/invest-go/proto"
)

type InstrumentAdapter struct {
	client *investgo.InstrumentsServiceClient
}

func NewInstrumentAdapter(client *investgo.InstrumentsServiceClient) *InstrumentAdapter {
	return &InstrumentAdapter{
		client: client,
	}
}

func (a *InstrumentAdapter) FetchBondCoupons(
	_ context.Context,
	bond instrument.BondRef,
	params instrument.FetchBondCouponParams,
) ([]instrument.BondCoupon, error) {
	resp, err := a.client.GetBondCoupons(bond.ID, params.From, params.To)
	if err != nil {
		return nil, err
	}

	ret := make([]instrument.BondCoupon, len(resp.Events))
	for i := range resp.Events {
		ret[i] = mapProtoCoupon(resp.Events[i])
	}

	return ret, nil
}

func mapProtoCoupon(c *proto.Coupon) instrument.BondCoupon {
	return instrument.BondCoupon{
		FIGI:             c.Figi,
		CouponDate:       c.CouponDate.AsTime(),
		CouponStartDate:  c.CouponStartDate.AsTime(),
		CouponEndDate:    c.CouponEndDate.AsTime(),
		CouponNumber:     int(c.CouponNumber),
		CouponPeriodDays: c.CouponPeriod,
		OneBondPay: instrument.Money{
			Whole:      c.PayOneBond.Units,
			Fractional: int64(c.PayOneBond.Nano),
		},
	}
}
