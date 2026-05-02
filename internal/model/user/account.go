package user

import "context"

type Account struct {
	ID     string
	Name   string
	Status AccountStatus
}

func (a *Account) Ref() Ref {
	return Ref{ID: a.ID}
}

type Ref struct {
	ID string
}

type AccountFilterer interface {
	FilterAccounts(ctx context.Context, p FilterParams) ([]Account, error)
}

type FilterParams struct {
	AccountStatus AccountStatus // If AccountStatusUnspecified - all accounts will be returned
}

type AccountStatus int

//go:generate go run github.com/dmarkham/enumer -type=AccountStatus -text -json -transform=lower -trimprefix=AccountStatus -output=status_enum.go
const (
	AccountStatusUnspecified AccountStatus = iota
	AccountStatusOpen
	AccountStatusClosed
	AccountStatusNew
)

type AccountRepository interface {
	AccountFilterer
}

type AccountRegistry struct {
	accounts AccountRepository
}

func NewAccountRegistry(accounts AccountRepository) *AccountRegistry {
	return &AccountRegistry{
		accounts: accounts,
	}
}

func (r *AccountRegistry) GetUserAccounts(ctx context.Context, params GetUserAccountParams) ([]Account, error) {
	return r.accounts.FilterAccounts(ctx, FilterParams{
		AccountStatus: params.Status,
	})
}

type GetUserAccountParams struct {
	Status AccountStatus
}
