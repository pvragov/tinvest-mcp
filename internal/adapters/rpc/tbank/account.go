package tbank

import (
	"context"
	"fmt"

	"github.com/pvragov/tinvest-mcp/internal/model/invest"

	"opensource.tbank.ru/invest/invest-go/investgo"
	proto "opensource.tbank.ru/invest/invest-go/proto"
)

type AccountAdapter struct {
	client *investgo.UsersServiceClient
}

func NewAccountAdapter(client *investgo.UsersServiceClient) *AccountAdapter {
	return &AccountAdapter{
		client: client,
	}
}

func (a *AccountAdapter) FilterAccounts(_ context.Context, params invest.FilterParams) ([]invest.Account, error) {
	resp, err := a.client.GetAccounts(new(mapAccountStatus[params.AccountStatus]))

	if err != nil {
		return nil, fmt.Errorf("failed to exec get accounts rpc: %w", err)
	}

	ret := make([]invest.Account, len(resp.Accounts))
	for i := range resp.Accounts {
		ret[i] = mapProtoAccount(resp.Accounts[i])
	}

	return ret, nil
}

var (
	mapAccountStatus = map[invest.AccountStatus]proto.AccountStatus{
		invest.AccountStatusOpen:        proto.AccountStatus_ACCOUNT_STATUS_OPEN,
		invest.AccountStatusClosed:      proto.AccountStatus_ACCOUNT_STATUS_CLOSED,
		invest.AccountStatusNew:         proto.AccountStatus_ACCOUNT_STATUS_NEW,
		invest.AccountStatusUnspecified: proto.AccountStatus_ACCOUNT_STATUS_ALL,
	}
	mapProtoAccountStatus = map[proto.AccountStatus]invest.AccountStatus{
		proto.AccountStatus_ACCOUNT_STATUS_OPEN:   invest.AccountStatusOpen,
		proto.AccountStatus_ACCOUNT_STATUS_CLOSED: invest.AccountStatusClosed,
		proto.AccountStatus_ACCOUNT_STATUS_NEW:    invest.AccountStatusNew,
	}
)

func mapProtoAccount(a *proto.Account) invest.Account {
	return invest.Account{
		ID:     a.Id,
		Name:   a.Name,
		Status: mapProtoAccountStatus[a.Status],
	}
}
