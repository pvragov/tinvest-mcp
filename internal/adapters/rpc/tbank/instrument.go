package tbank

import (
	"context"
	"fmt"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"

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
		return nil, fmt.Errorf("failed to exec get bond coupons rpc: %w", err)
	}

	ret := make([]instrument.BondCoupon, len(resp.Events))
	for i := range resp.Events {
		ret[i] = mapProtoCoupon(resp.Events[i])
	}

	return ret, nil
}

func mapProtoCoupon(c *proto.Coupon) instrument.BondCoupon {
	return instrument.BondCoupon{
		CouponDate:       c.CouponDate.AsTime(),
		CouponNumber:     int(c.CouponNumber),
		CouponPeriodDays: c.CouponPeriod,
		OneBondPay: instrument.Money{
			Units:      c.PayOneBond.Units,
			MinorUnits: c.PayOneBond.Nano / 1_000_0000,
			Currency:   c.PayOneBond.Currency,
		},
	}
}

func (a *InstrumentAdapter) FetchBond(_ context.Context, bond *instrument.Bond) error {
	resp, err := a.client.BondByUid(bond.ID)
	if err != nil {
		return fmt.Errorf("failed to exec bond by uid rpc: %w", err)
	}

	in := resp.GetInstrument()
	*bond = instrument.Bond{
		ID:                bond.ID,
		Name:              in.GetName(),
		ISIN:              in.GetIsin(),
		Currency:          in.GetCurrency(),
		HasAmortization:   in.GetAmortizationFlag(),
		HasFloatingCoupon: in.GetFloatingCouponFlag(),
	}

	return nil
}

func (a *InstrumentAdapter) FetchShare(_ context.Context, share *instrument.Share) error {
	resp, err := a.client.ShareByUid(share.ID)
	if err != nil {
		return fmt.Errorf("failed to exec share by uid rpc: %w", err)
	}

	in := resp.GetInstrument()
	*share = instrument.Share{
		ID:       share.ID,
		Name:     in.GetName(),
		ISIN:     in.GetIsin(),
		Currency: in.GetCurrency(),
	}

	return nil
}
