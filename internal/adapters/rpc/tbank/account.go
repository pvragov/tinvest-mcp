package tbank

import (
	"context"
	"fmt"

	"tinkoff-invest-mcp/internal/model/user"

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

func (a *AccountAdapter) FilterAccounts(_ context.Context, params user.FilterParams) ([]user.Account, error) {
	resp, err := a.client.GetAccounts(new(mapAccountStatus[params.AccountStatus]))

	if err != nil {
		return nil, fmt.Errorf("failed to exec get accounts rpc: %w", err)
	}

	ret := make([]user.Account, len(resp.Accounts))
	for i := range resp.Accounts {
		ret[i] = mapProtoAccount(resp.Accounts[i])
	}

	return ret, nil
}

var (
	mapAccountStatus = map[user.AccountStatus]proto.AccountStatus{
		user.AccountStatusOpen:        proto.AccountStatus_ACCOUNT_STATUS_OPEN,
		user.AccountStatusClosed:      proto.AccountStatus_ACCOUNT_STATUS_CLOSED,
		user.AccountStatusNew:         proto.AccountStatus_ACCOUNT_STATUS_NEW,
		user.AccountStatusUnspecified: proto.AccountStatus_ACCOUNT_STATUS_ALL,
	}
	mapProtoAccountStatus = map[proto.AccountStatus]user.AccountStatus{
		proto.AccountStatus_ACCOUNT_STATUS_OPEN:   user.AccountStatusOpen,
		proto.AccountStatus_ACCOUNT_STATUS_CLOSED: user.AccountStatusClosed,
		proto.AccountStatus_ACCOUNT_STATUS_NEW:    user.AccountStatusNew,
	}
)

func mapProtoAccount(a *proto.Account) user.Account {
	return user.Account{
		ID:     a.Id,
		Name:   a.Name,
		Status: mapProtoAccountStatus[a.Status],
	}
}
