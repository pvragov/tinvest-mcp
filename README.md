# TBANK INVEST MCP SERVER

MCP сервер для работы с API tbank инвестиций

## Установка 

```shell
go install github.com/pvragov/tinvest-mcp/cmd/tinvest-mcp@latest
```

## Конфигурирование

Сервер конфигурируется через переменные окружения

```shell
TBANK_INVEST_MCP_SERVER_LISTEN="0.0.0.0:32900"
TBANK_INVEST_API_TOKEN=""
TBANK_INVEST_API_ENDPOINT="invest-public-api.tbank.ru:443"
```