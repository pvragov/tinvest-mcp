# TBANK INVEST MCP SERVER

MCP сервер для работы с API tbank инвестиций

## Установка 

```shell
go install github.com/pvragov/tinvest-mcp/cmd/tinvest-mcp@latest
```

## Конфигурирование

Сервер конфигурируется через переменные окружения

| Имя переменной                    | Обязательность | Описание                                          |
|-----------------------------------|----------------|---------------------------------------------------|
| TBANK_INVEST_MCP_API_TOKEN        | +              | API токен tbank инвестиций                        |
| TBANK_INVEST_MCP_API_ENDPOINT     | -              | API endpoint tbank инвестиций в формате host:port |
| TBANK_INVEST_MCP_TLS_CA_CERT_PATH | -              | Абсолютный путь до CA сертификата                 |
| TBANK_INVEST_MCP_TLS_SKIP_VERIFY  | -              | Пропустить проверку сертификата сервера           |

Для работы с API инвестиций необходимо установить сертификаты **НУЦ Минцифры РФ**.  
Их можно скачать на сайте [Госуслуг](https://www.gosuslugi.ru/crt), либо взять сертификат `certs/tbank_ca.pem` из этого репозитория и
указать абсолютный путь к нему в переменной окружения `TBANK_INVEST_MCP_TLS_CA_CERT_PATH`.

Если установка сертификата невозможна, проверку можно отключить, задав `TBANK_INVEST_MCP_TLS_SKIP_VERIFY=true` (крайне не рекомендуется).

## Запуск

Сервер запускается командой:

```shell
tinvest-mcp
```

В качестве транспорта используется stdin
