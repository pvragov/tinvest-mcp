package instrument

import (
	"context"
	"fmt"
	"time"
)

type BondCouponsFetcher interface {
	FetchBondCoupons(ctx context.Context, bond BondRef, params FetchBondCouponParams) ([]BondCoupon, error)
}

type FetchBondCouponParams struct {
	From time.Time
	To   time.Time
}

type BondCoupon struct {
	FIGI             string
	CouponDate       time.Time
	CouponStartDate  time.Time
	CouponEndDate    time.Time
	CouponNumber     int
	CouponPeriodDays int32
	OneBondPay       Money
}

type Money struct {
	IntPart     int64
	DecimalPart int32
}

func (m *Money) String() string {
	if m.IntPart >= 0 {
		return fmt.Sprintf("%d.%d", m.IntPart, m.DecimalPart)
	}

	return fmt.Sprintf("-%d.%d", m.IntPart, m.DecimalPart)
}

type Repository interface {
	BondCouponsFetcher
}

type BondRegistry struct {
	bonds Repository
}

func NewBondRegistry(bonds Repository) *BondRegistry {
	return &BondRegistry{
		bonds: bonds,
	}
}

func (r *BondRegistry) GetBondCoupons(ctx context.Context, bond BondRef, params GetBondCouponsParams) ([]BondCoupon, error) {
	return r.bonds.FetchBondCoupons(ctx, bond, FetchBondCouponParams{
		From: params.From,
		To:   params.To,
	})
}

type GetBondCouponsParams struct {
	From time.Time
	To   time.Time
}

type BondRef struct {
	ID string
}
