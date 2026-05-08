package tbank

import (
	"context"
	"fmt"

	"github.com/pvragov/tinvest-mcp/internal/model/instrument"
	"github.com/sevlyar/box"

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
) ([]instrument.Coupon, error) {
	resp, err := a.client.GetBondCoupons(bond.ID, params.From, params.To)
	if err != nil {
		return nil, fmt.Errorf("failed to exec get bond coupons rpc: %w", err)
	}

	ret := make([]instrument.Coupon, len(resp.Events))
	for i := range resp.Events {
		ret[i] = mapProtoCoupon(resp.Events[i])
	}

	return ret, nil
}

func mapProtoCoupon(c *proto.Coupon) instrument.Coupon {
	ret := instrument.Coupon{
		No:         int(c.CouponNumber),
		PeriodDays: c.CouponPeriod,
		OneBondPay: mapProtoMoney(c.GetPayOneBond()),
	}

	if c.CouponDate != nil {
		ret.PayDate = box.Some(c.CouponDate.AsTime())
	}

	if c.CouponStartDate != nil && c.CouponEndDate != nil {
		ret.Period = box.Some(instrument.CuponPeriod{
			Start: c.GetCouponStartDate().AsTime(),
			End:   c.GetCouponEndDate().AsTime(),
		})
	}

	return ret
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
		Nominal:           mapProtoMoney(in.GetNominal()),
		InitialNominal:    mapProtoMoney(in.GetInitialNominal()),
		HasAmortization:   in.GetAmortizationFlag(),
		HasFloatingCoupon: in.GetFloatingCouponFlag(),
		LotSize:           int(in.GetLot()),
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
		LotSize:  int(in.GetLot()),
	}

	return nil
}

func (a *InstrumentAdapter) FetchShareDividends(
	_ context.Context,
	ref instrument.ShareRef,
	params instrument.FetchShareDividendsParams,
) ([]instrument.Dividend, error) {
	resp, err := a.client.GetDividents(ref.ID, params.From, params.To)
	if err != nil {
		return nil, fmt.Errorf("failed to exec get dividends rpc: %w", err)
	}

	ret := make([]instrument.Dividend, len(resp.Dividends))
	for i := range resp.Dividends {
		ret[i] = mapProtoDividend(resp.Dividends[i])
	}

	return ret, nil
}

func mapProtoDividend(c *proto.Dividend) instrument.Dividend {
	divNet := c.GetDividendNet()
	yield := c.GetYieldValue()

	return instrument.Dividend{
		Value:        mapProtoMoney(divNet),
		PaymentDate:  c.GetPaymentDate().AsTime(),
		DeclaredDate: c.GetDeclaredDate().AsTime(),
		LastBuyDate:  c.GetLastBuyDate().AsTime(),
		YieldValue:   float64(yield.GetUnits()) + float64(yield.GetNano())/1_000_000_000,
	}
}

func mapProtoMoney(v *proto.MoneyValue) instrument.Money {
	return instrument.Money{
		Units:      v.GetUnits(),
		MinorUnits: v.GetNano() / 10_000_000,
		Currency:   v.GetCurrency(),
	}
}
