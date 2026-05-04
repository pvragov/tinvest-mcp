package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/netip"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/mark3labs/mcp-go/server"
	"github.com/pvragov/tinvest-mcp/internal/adapters/mcp"
	"github.com/pvragov/tinvest-mcp/internal/adapters/rpc/tbank"
	"github.com/pvragov/tinvest-mcp/internal/model/instrument"
	"github.com/pvragov/tinvest-mcp/internal/model/invest"
	"opensource.tbank.ru/invest/invest-go/investgo"
)

func main() {
	if err := run(); err != nil {
		slog.Error("failed to run mcp server", "error", err)
		os.Exit(1)
	}
}

func run() error {
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
		return fmt.Errorf("failed to parse tbank client config: %v", err)
	}

	client, err := newTBankClient(ctx, clientConfig)
	if err != nil {
		return fmt.Errorf("failed to create tbank client: %v", err)
	}

	var (
		accountAdapter   = tbank.NewAccountAdapter(client.NewUsersServiceClient())
		portfolioAdapter = tbank.NewPortfolioAdapter(client.NewOperationsServiceClient())
		bondsAdapter     = tbank.NewInstrumentAdapter(client.NewInstrumentsServiceClient())
	)

	s.AddTools(
		mcp.NewGetUserAccountsTool(invest.NewAccountRegistry(accountAdapter)),
		mcp.NewGetPortfolio(invest.NewPortfolioRegistry(portfolioAdapter)),
		mcp.NewGetBondCouponsTool(instrument.NewBondRegistry(bondsAdapter)),
	)

	var httpDebugServerEnable bool // use http only for debug
	flag.BoolVar(&httpDebugServerEnable, "http", false, "run mcp over http")
	flag.Parse()

	errChan := make(chan error)

	go func() {
		var serveErr error
		if httpDebugServerEnable {
			serveErr = serveHTTP(s)
		} else {
			server.ServeStdio(s)
		}

		if serveErr != nil {
			errChan <- serveErr
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return fmt.Errorf("serve error: %w", err)
	}
}

func serveHTTP(s *server.MCPServer) error {
	config, err := parseHTTPServerConfig()
	if err != nil {
		return fmt.Errorf("failed to parse http server config: %v", err)
	}

	slog.Info("starting http server", "listen", config.Listen)

	return server.NewStreamableHTTPServer(s).Start(config.Listen)
}

func newTBankClient(ctx context.Context, config investgo.Config) (*investgo.Client, error) {
	client, err := investgo.NewClient(ctx, config, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func parseTBankClientConfig() (investgo.Config, error) {
	token := os.Getenv("TBANK_INVEST_MCP_API_TOKEN")
	if token == "" {
		return investgo.Config{}, fmt.Errorf("tbank token must not be empty")
	}

	endpoint := os.Getenv("TBANK_INVEST_MCP_API_ENDPOINT")
	if endpoint == "" {
		endpoint = "invest-public-api.tbank.ru:443"
	}

	var (
		skipVerify bool
		err        error
	)
	if skipVerifyParam := os.Getenv("TBANK_INVEST_MCP_TLS_SKIP_VERIFY"); skipVerifyParam != "" {
		skipVerify, err = strconv.ParseBool(skipVerifyParam)
		if err != nil {
			return investgo.Config{}, fmt.Errorf("failed to parse skip verify param: %v", err)
		}
	}

	const appName = "tbank-mcp"

	return investgo.Config{
		EndPoint:           endpoint,
		Token:              token,
		AppName:            appName,
		TLSCACertFile:      os.Getenv("TBANK_INVEST_MCP_TLS_CA_CERT_PATH"),
		InsecureSkipVerify: skipVerify,
	}, nil
}

func parseHTTPServerConfig() (httpServerConfig, error) {
	addr, err := netip.ParseAddrPort(os.Getenv("TBANK_INVEST_MCP_SERVER_LISTEN"))
	if err != nil {
		return httpServerConfig{}, fmt.Errorf("failed to parse listen addr: %v", err)
	}

	if addr.Port() == 0 {
		return httpServerConfig{}, fmt.Errorf("port must not be 0")
	}

	return httpServerConfig{
		Listen: addr.String(),
	}, nil
}

type httpServerConfig struct {
	Listen string
}
