# TBANK INVEST MCP SERVER

MCP сервер для работы с API tbank инвестиций

## Установка 

```shell
go install github.com/pvragov/tinvest-mcp/cmd/tinvest-mcp@latest
```

## Конфигурирование

Сервер конфигурируется через переменные окружения

| Имя переменной            | Обязательность | Описание                       |
|---------------------------|----------------|--------------------------------|
| TBANK_INVEST_API_TOKEN    | +              | API токен tbank инвестиций     |
| TBANK_INVEST_API_ENDPOINT | -              | API endpoint tbank инвестиций  |

## Запуск

Сервер запускается командой:

```shell
tinvest-mcp
```

В качестве транспорта используется stdin
