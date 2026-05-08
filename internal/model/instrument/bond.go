package instrument

import (
	"context"
	"fmt"
	"time"
)

type BondCouponsFetcher interface {
	// FetchBondCoupons fetches the bound coupons by the specified params.
	FetchBondCoupons(ctx context.Context, bond BondRef, params FetchBondCouponParams) ([]BondCoupon, error)
}

type FetchBondCouponParams struct {
	From time.Time
	To   time.Time
}

type BondFetcher interface {
	// FetchBond fetches bond by id.
	FetchBond(ctx context.Context, bond *Bond) error
}

type Bond struct {
	ID                string // Instrument ID
	Name              string
	ISIN              string
	Currency          string
	LotSize           int
	Nominal           Money
	InitialNominal    Money
	HasAmortization   bool
	HasFloatingCoupon bool
}

type BondRef struct {
	ID string // Instrument ID
}

type BondCoupon struct {
	CouponDate       time.Time
	CouponNumber     int
	CouponPeriodDays int32
	OneBondPay       Money
}

type Money struct {
	Units      int64
	MinorUnits int32
	Currency   string
}

func (m *Money) String() string {
	if m.MinorUnits < 0 {
		return fmt.Sprintf("%d.%d", m.Units, m.MinorUnits*-1)
	}

	return fmt.Sprintf("%d.%d", m.Units, m.MinorUnits)
}

type Repository interface {
	BondFetcher
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

func (r *BondRegistry) GetBondCoupons(
	ctx context.Context, bond BondRef, params GetBondCouponsParams,
) ([]BondCoupon, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	return r.bonds.FetchBondCoupons(ctx, bond, FetchBondCouponParams(params))
}

type GetBondCouponsParams struct {
	From time.Time
	To   time.Time
}

func (p *GetBondCouponsParams) Validate() error {
	if p.From.After(p.To) {
		return fmt.Errorf("from time must be before to time")
	}

	return nil
}

func (r *BondRegistry) GetBond(ctx context.Context, ref BondRef) (*Bond, error) {
	bond := &Bond{ID: ref.ID}
	if err := r.bonds.FetchBond(ctx, bond); err != nil {
		return nil, fmt.Errorf("failed to fetch bond: %w", err)
	}

	return bond, nil
}
