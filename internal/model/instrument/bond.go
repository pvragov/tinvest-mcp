package instrument

import (
	"context"
	"time"

	"github.com/sevlyar/box"
)

type BondFetcher interface {
	// FetchBond fetches bond by id.
	FetchBond(ctx context.Context, bond *Bond) error
}

type BondCouponsFetcher interface {
	// FetchBondCoupons fetches the bound coupons by the specified params.
	FetchBondCoupons(ctx context.Context, bond BondRef, params FetchBondCouponParams) ([]BondCoupon, error)
}

type FetchBondCouponParams struct {
	From time.Time
	To   time.Time
}

type BondRedemptionFetcher interface {
	// FetchBondRedemptions fetches the bond redemptions by the specified bond.
	FetchBondRedemptions(ctx context.Context, bond BondRef, params FetchBondRedemptionParams) ([]BondRedemption, error)
}

type FetchBondRedemptionParams struct {
	From time.Time
	To   time.Time
}

type BondRepository interface {
	BondFetcher
	BondCouponsFetcher
	BondRedemptionFetcher
}

type Bond struct {
	ID                string // Instrument ID
	Name              string
	Ticker            string
	ClassCode         string
	ISIN              string
	Currency          string
	LotSize           int
	Nominal           Money
	InitialNominal    Money
	MaturityDate      box.Optional[time.Time]
	ACI               box.Optional[Money]
	HasAmortization   bool
	HasFloatingCoupon bool
}

type BondRef struct {
	ID string // Instrument ID
}

type BondCoupon struct {
	PayDate    box.Optional[time.Time]
	Period     box.Optional[CuponPeriod]
	PeriodDays int32
	No         int
	OneBondPay Money
}

type BondRedemption struct {
	PayDate    time.Time
	OneBondPay Money
}

type CuponPeriod struct {
	Start time.Time
	End   time.Time
}
