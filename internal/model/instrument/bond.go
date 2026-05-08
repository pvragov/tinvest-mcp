package instrument

import (
	"context"
	"fmt"
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

type Repository interface {
	BondFetcher
	BondCouponsFetcher
	BondRedemptionFetcher
}

type Bond struct {
	ID                string // Instrument ID
	Name              string
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

func (r *BondRegistry) GetBondRedemptions(
	ctx context.Context,
	bond BondRef,
	params GetBondRedemptionParams,
) ([]BondRedemption, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	redemptions, err := r.bonds.FetchBondRedemptions(ctx, bond, FetchBondRedemptionParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bond redemptions: %w", err)
	}

	return redemptions, nil
}

type GetBondRedemptionParams struct {
	From time.Time
	To   time.Time
}

func (p *GetBondRedemptionParams) Validate() error {
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
