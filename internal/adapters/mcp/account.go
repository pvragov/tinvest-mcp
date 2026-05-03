package mcp

import (
	"context"
	"fmt"

	"github.com/pvragov/tinvest-mcp/internal/model/invest"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type UserAccountService interface {
	GetUserAccounts(ctx context.Context, params invest.GetUserAccountParams) ([]invest.Account, error)
}

func NewGetUserAccountsTool(service UserAccountService) server.ServerTool {
	const accountStatusArgName = "status"

	return server.ServerTool{
		Tool: mcp.NewTool(
			"get-user-accounts",
			mcp.WithDescription("Позволяет получить список счетов пользователя"),
			mcp.WithString("status", mcp.Enum("open", "closed", "new")),
			mcp.WithOutputSchema[GetUserAccountToolReply](),
		),
		Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			statusArg := req.GetString(accountStatusArgName, "")

			var status invest.AccountStatus
			if statusArg != "" {
				var err error
				status, err = invest.AccountStatusString(statusArg)
				if err != nil {
					return nil, fmt.Errorf("invalid account status: %w", err)
				}
			}

			accounts, err := service.GetUserAccounts(ctx, invest.GetUserAccountParams{
				Status: status,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get user accounts: %w", err)
			}

			resp := GetUserAccountToolReply{
				Accounts: make([]AccountView, len(accounts)),
			}
			for i, account := range accounts {
				resp.Accounts[i] = mapUserAccount(account)
			}

			return mcp.NewToolResultJSON(resp)
		},
	}
}

type GetUserAccountToolReply struct {
	Accounts []AccountView `json:"accounts"`
}

type AccountView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func mapUserAccount(a invest.Account) AccountView {
	return AccountView{
		ID:   a.ID,
		Name: a.Name,
	}
}
