package instrument

import (
	"context"
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

type ShareRef struct {
	ID string
}
