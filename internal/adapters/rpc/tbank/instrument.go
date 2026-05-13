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
	bond instrument.Ref,
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
	ret := instrument.BondCoupon{
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

func (a *InstrumentAdapter) FetchBondRedemptions(
	_ context.Context,
	bond instrument.Ref,
	params instrument.FetchBondRedemptionParams,
) ([]instrument.BondRedemption, error) {
	resp, err := a.client.GetBondEvents(bond.ID, proto.GetBondEventsRequest_EVENT_TYPE_MTY, params.From, params.To)
	if err != nil {
		return nil, fmt.Errorf("failed to exec get bond events rpc: %w", err)
	}

	ret := make([]instrument.BondRedemption, len(resp.Events))
	for i := range ret {
		ret[i] = mapProtoBondRedemptionEvent(resp.Events[i])
	}

	return ret, nil
}

func mapProtoBondRedemptionEvent(e *proto.GetBondEventsResponse_BondEvent) instrument.BondRedemption {
	return instrument.BondRedemption{
		PayDate:    e.GetPayDate().AsTime(),
		OneBondPay: mapProtoMoney(e.GetPayOneBond()),
	}
}

func (a *InstrumentAdapter) FetchBond(_ context.Context, bond *instrument.Bond) error {
	resp, err := a.client.BondByUid(bond.ID)
	if err != nil {
		return fmt.Errorf("failed to exec bond by uid rpc: %w", err)
	}

	in := resp.GetInstrument()
	*bond = instrument.Bond{
		Instrument: instrument.Instrument{
			Type:      instrument.TypeBond,
			ID:        in.GetUid(),
			Ticker:    in.GetTicker(),
			ClassCode: in.GetClassCode(),
		},
		Name:              in.GetName(),
		Ticker:            in.GetTicker(),
		ClassCode:         in.GetClassCode(),
		ISIN:              in.GetIsin(),
		Currency:          in.GetCurrency(),
		Nominal:           mapProtoMoney(in.GetNominal()),
		InitialNominal:    mapProtoMoney(in.GetInitialNominal()),
		HasAmortization:   in.GetAmortizationFlag(),
		HasFloatingCoupon: in.GetFloatingCouponFlag(),
		LotSize:           int(in.GetLot()),
	}

	// TODO: add optional type mapping helper
	if in.MaturityDate != nil {
		bond.MaturityDate = box.Some(in.MaturityDate.AsTime())
	}

	if in.AciValue != nil {
		bond.ACI = box.Some(mapProtoMoney(in.GetAciValue()))
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
		Instrument: instrument.Instrument{
			Type:      instrument.TypeShare,
			ID:        in.GetUid(),
			Ticker:    in.GetTicker(),
			ClassCode: in.GetClassCode(),
		},
		Name:     in.GetName(),
		ISIN:     in.GetIsin(),
		Currency: in.GetCurrency(),
		LotSize:  int(in.GetLot()),
	}

	return nil
}

func (a *InstrumentAdapter) FetchShareDividends(
	_ context.Context,
	ref instrument.Ref,
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

func (a *InstrumentAdapter) SearchInstrument(_ context.Context, query string) ([]instrument.Info, error) {
	resp, err := a.client.FindInstrument(query)
	if err != nil {
		return nil, fmt.Errorf("failed to exec find instrument rpc: %w", err)
	}

	ret := make([]instrument.Info, len(resp.Instruments))
	for i, in := range resp.Instruments {
		ret[i] = instrument.Info{
			Instrument: instrument.Instrument{
				ID:        in.GetUid(),
				Type:      mapInstrumentType[in.GetInstrumentType()],
				Ticker:    in.GetTicker(),
				ClassCode: in.GetClassCode(),
			},
			ISIN: in.GetIsin(),
			Name: in.GetName(),
		}
	}

	return ret, nil
}

func (a *InstrumentAdapter) FetchETF(_ context.Context, _ *instrument.ETF) error {
	panic("implement me")
}
