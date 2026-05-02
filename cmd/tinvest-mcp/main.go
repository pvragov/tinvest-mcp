package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"os"
	"os/signal"
	"syscall"

	"github.com/pvragov/tinvest-mcp/internal/adapters/mcp"
	"github.com/pvragov/tinvest-mcp/internal/adapters/rpc/tbank"
	"github.com/pvragov/tinvest-mcp/internal/model/instrument"
	"github.com/pvragov/tinvest-mcp/internal/model/portfolio"
	"github.com/pvragov/tinvest-mcp/internal/model/user"

	"github.com/mark3labs/mcp-go/server"
	"opensource.tbank.ru/invest/invest-go/investgo"
)

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"TBank mcp server",
		"0.1.0",
		server.WithToolCapabilities(false),
		server.WithLogging(),
		server.WithRecovery(),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()

	clientConfig, err := parseTBankClientConfig()
	if err != nil {
		slog.Error("failed to parse tbank client config", "error", err)
		os.Exit(1)
	}

	client, err := NewTBankClient(ctx, clientConfig)
	if err != nil {
		slog.Error("failed to create tbank client: %v", "error", err)
		os.Exit(1)
	}

	serverConfig, err := parseMCPServerConfig()
	if err != nil {
		slog.Error("failed to parse mcp server config", "error", err)
		os.Exit(1)
	}

	var (
		accountAdapter   = tbank.NewAccountAdapter(client.NewUsersServiceClient())
		portfolioAdapter = tbank.NewPortfolioAdapter(client.NewOperationsServiceClient())
		bondsAdapter     = tbank.NewInstrumentAdapter(client.NewInstrumentsServiceClient())
	)

	s.AddTools(
		mcp.NewGetUserAccountsTool(user.NewAccountRegistry(accountAdapter)),
		mcp.NewGetPortfolio(portfolio.NewRegistry(portfolioAdapter)),
		mcp.NewGetBondCouponsTool(instrument.NewBondRegistry(bondsAdapter)),
	)

	errChan := make(chan error)

	go func() {
		if err := server.NewStreamableHTTPServer(s).Start(serverConfig.Listen); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		os.Exit(0)
	case err := <-errChan:
		fmt.Printf("serve error: %v\n", err)
	}
}

func NewTBankClient(ctx context.Context, config investgo.Config) (*investgo.Client, error) {
	client, err := investgo.NewClient(ctx, investgo.Config{
		EndPoint:           "invest-public-api.tbank.ru:443",
		Token:              "test",
		AppName:            "tinkoff-mcp",
		InsecureSkipVerify: true,
	}, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func parseTBankClientConfig() (investgo.Config, error) {
	token := os.Getenv("TBANK_INVEST_API_TOKEN")
	if token == "" {
		return investgo.Config{}, fmt.Errorf("tbank token must not be empty")
	}

	endpoint := os.Getenv("TBANK_INVEST_API_ENDPOINT")
	if endpoint == "" {
		return investgo.Config{}, fmt.Errorf("tbank endpoint must not be empty")
	}

	const appName = "tbank-mcp"

	return investgo.Config{
		EndPoint: endpoint,
		Token:    token,
		AppName:  appName,
	}, nil
}

func parseMCPServerConfig() (mcpServerConfig, error) {
	addr, err := netip.ParseAddrPort(os.Getenv("TBANK_INVEST_MCP_SERVER_LISTEN"))
	if err != nil {
		return mcpServerConfig{}, fmt.Errorf("failed to parse listen addr: %v", err)
	}

	if addr.Port() == 0 {
		return mcpServerConfig{}, fmt.Errorf("port must not be 0")
	}

	return mcpServerConfig{
		Listen: addr.String(),
	}, nil
}

type mcpServerConfig struct {
	Listen string
}
