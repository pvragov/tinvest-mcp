package instrument

import (
	"context"
	"fmt"
	"time"
)

type Type int

//go:generate go run github.com/dmarkham/enumer -type=Type -text -json -yaml -transform=lower -trimprefix=Type -output=type_enum.go
const (
	TypeUnknown Type = iota
	TypeShare
	TypeBond
	TypeCurrency
	TypeETF
)

type Instrument struct {
	Type      Type
	ID        string
	Ticker    string
	ClassCode string
}

type Ref struct {
	ID string
}

type Info struct {
	Instrument
	ISIN string
	Name string
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

type Searcher interface {
	SearchInstrument(ctx context.Context, query string) ([]Info, error)
}

type Repository interface {
	BondRepository
	ShareRepository
	Searcher
}

type Registry struct {
	repo Repository
}

func NewRegistry(repo Repository) *Registry {
	return &Registry{
		repo: repo,
	}
}

func (r *Registry) SearchInstrument(ctx context.Context, query string) ([]Info, error) {
	const queryMinLen = 3
	if query == "" || len(query) < queryMinLen {
		return nil, fmt.Errorf("invalid query")
	}

	return r.repo.SearchInstrument(ctx, query)
}

func (r *Registry) GetBondCoupons(
	ctx context.Context, bond Ref, params GetBondCouponsParams,
) ([]BondCoupon, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	return r.repo.FetchBondCoupons(ctx, bond, FetchBondCouponParams(params))
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

func (r *Registry) GetBondRedemptions(
	ctx context.Context,
	bond Ref,
	params GetBondRedemptionParams,
) ([]BondRedemption, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	redemptions, err := r.repo.FetchBondRedemptions(ctx, bond, FetchBondRedemptionParams(params))
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

func (r *Registry) GetBond(ctx context.Context, ref Ref) (*Bond, error) {
	bond := NewBond(ref.ID)
	if err := r.repo.FetchBond(ctx, bond); err != nil {
		return nil, fmt.Errorf("failed to fetch bond: %w", err)
	}

	return bond, nil
}

func (r *Registry) GetShare(ctx context.Context, ref Ref) (*Share, error) {
	share := NewShare(ref.ID)
	if err := r.repo.FetchShare(ctx, share); err != nil {
		return nil, fmt.Errorf("failed to fetch share: %w", err)
	}

	return share, nil
}

func (r *Registry) GetShareDividends(
	ctx context.Context, share Ref, params GetShareDividendsParams,
) ([]Dividend, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	dividends, err := r.repo.FetchShareDividends(ctx, share, FetchShareDividendsParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch share dividends: %w", err)
	}

	return dividends, nil
}

type GetShareDividendsParams struct {
	From time.Time
	To   time.Time
}

func (p *GetShareDividendsParams) Validate() error {
	if p.From.After(p.To) {
		return fmt.Errorf("from time must be before to time")
	}

	return nil
}
