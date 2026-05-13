package instrument

import "context"

type ETFFetcher interface {
	// FetchETF fetches the ETF by the specified instrument ID.
	FetchETF(ctx context.Context, etf *ETF) error
}

type ETFRepository interface {
	ETFFetcher
}

type ETF struct {
	Instrument
	Name     string
	ISIN     string
	Currency string
	LotSize  int
}

func NewETF(id string) *ETF {
	return &ETF{
		Instrument: Instrument{
			ID: id,
		},
	}
}
