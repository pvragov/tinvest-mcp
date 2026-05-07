package mcp

import (
	"context"
	"testing"

	"github.com/pvragov/tinvest-mcp/internal/model/invest"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewGetUserAccountsTool(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		account := invest.Account{
			ID:   "123",
			Name: "Test account",
		}

		service := &MockUserAccountService{}
		service.On("GetUserAccounts", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			params := args.Get(1).(invest.GetUserAccountParams)
			require.Equal(t, invest.AccountStatusOpen, params.Status)
		}).Return([]invest.Account{account}, nil)

		tool := NewGetUserAccountsTool(service)

		res, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-user-accounts"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"status": "open"},
			},
		})

		require.NoError(t, err)
		expected := getUserAccountToolReply{
			Accounts: []accountView{{ID: "123", Name: "Test account"}},
		}
		require.Equal(t, expected, res.StructuredContent)
	})

	t.Run("invalid status", func(t *testing.T) {
		tool := NewGetUserAccountsTool(NewMockUserAccountServiceStub())

		_, err := tool.Handler(t.Context(), mcp.CallToolRequest{
			Request: mcp.Request{Method: "tbank-get-user-accounts"},
			Params: mcp.CallToolParams{
				Arguments: map[string]interface{}{"status": "invalid"},
			},
		})

		require.Error(t, err)
	})

}

type MockUserAccountService struct {
	mock.Mock
}

func NewMockUserAccountServiceStub() *MockUserAccountService {
	s := &MockUserAccountService{}
	s.On("GetUserAccounts", mock.Anything, mock.Anything).Return([]invest.Account{}, nil)
	return s
}

func (m *MockUserAccountService) GetUserAccounts(ctx context.Context, params invest.GetUserAccountParams) ([]invest.Account, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]invest.Account), args.Error(1)
}
