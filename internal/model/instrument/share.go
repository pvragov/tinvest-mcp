package instrument

import (
	"context"
	"fmt"
)

type ShareFetcher interface {
	// FetchShare fetches the share by the specified ID.
	FetchShare(ctx context.Context, share *Share) error
}

type ShareRepository interface {
	ShareFetcher
}

type Share struct {
	ID       string
	Name     string
	ISIN     string
	Currency string
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
