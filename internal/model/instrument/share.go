package instrument

import (
	"context"
	"fmt"
	"time"
)

type ShareFetcher interface {
	// FetchShare fetches the share by the specified ID.
	FetchShare(ctx context.Context, share *Share) error
}

type ShareDividendFetcher interface {
	// FetchShareDividends fetches the share dividends by the specified ID.
	FetchShareDividends(ctx context.Context, ref ShareRef, params FetchShareDividendsParams) ([]Dividend, error)
}

type FetchShareDividendsParams struct {
	From time.Time
	To   time.Time
}

type ShareRepository interface {
	ShareFetcher
	ShareDividendFetcher
}

type Share struct {
	ID       string
	Name     string
	ISIN     string
	Currency string
	LotSize  int
}

type Dividend struct {
	Value        Money
	PaymentDate  time.Time
	DeclaredDate time.Time // Дата объявления
	LastBuyDate  time.Time // Последний день (включительно) покупки для получения выплаты по UTC.
	YieldValue   float64   // Величина доходности в процентах
}

type ShareRegistry struct {
	shares ShareRepository
}

func NewShareRegistry(shares ShareRepository) *ShareRegistry {
	return &ShareRegistry{
		shares: shares,
	}
}

type ShareRef struct {
	ID string
}

func (r *ShareRegistry) GetShare(ctx context.Context, ref ShareRef) (*Share, error) {
	share := &Share{ID: ref.ID}
	if err := r.shares.FetchShare(ctx, share); err != nil {
		return nil, fmt.Errorf("failed to fetch share: %w", err)
	}

	return share, nil
}

func (r *ShareRegistry) GetShareDividends(
	ctx context.Context, share ShareRef, params GetShareDividendsParams,
) ([]Dividend, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	dividends, err := r.shares.FetchShareDividends(ctx, share, FetchShareDividendsParams(params))
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
