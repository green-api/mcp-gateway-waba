# mcp-gateway-waba

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/green-api/mcp-gateway-waba/main)](https://go.dev/)
[![release](https://img.shields.io/github/v/release/green-api/mcp-gateway-waba)](https://github.com/green-api/mcp-gateway-waba/releases)
[![Protocol: MCP](https://img.shields.io/badge/Protocol-MCP-orange.svg)](https://modelcontextprotocol.io)
[![Go Report Card](https://goreportcard.com/badge/github.com/green-api/mcp-gateway-waba)](https://goreportcard.com/report/github.com/green-api/mcp-gateway-waba)

- [Documentation in English](README.md)

## Поддержка

[![Support](https://img.shields.io/badge/support@green--api.com-D14836?style=for-the-badge&logo=gmail&logoColor=white)](mailto:support@green-api.com)
[![Support](https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white)](https://t.me/greenapi_support_ru_bot)
[![Support](https://img.shields.io/badge/WhatsApp-25D366?style=for-the-badge&logo=whatsapp&logoColor=white)](https://wa.me/77780739095)

## Руководства и новости

[![Guides](https://img.shields.io/badge/YouTube-%23FF0000.svg?style=for-the-badge&logo=YouTube&logoColor=white)](https://www.youtube.com/@green-api)
[![News](https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white)](https://t.me/green_api)
[![News](https://img.shields.io/badge/WhatsApp-25D366?style=for-the-badge&logo=whatsapp&logoColor=white)](https://whatsapp.com/channel/0029VaHUM5TBA1f7cG29nO1C)

---

MCP-шлюз для GREEN-API **WABA**, официального WhatsApp Business API. Позволяет агентам ИИ (Claude Desktop, ChatGPT, Cursor, OpenClaw и другие) отправлять шаблонные и сессионные сообщения, управлять шаблонами сообщений и получать уведомления от аккаунта WhatsApp Business, используя [Model Context Protocol](https://modelcontextprotocol.io).

Это WABA-версия шлюза [green-api-mcp-gateway](https://github.com/green-api/green-api-mcp-gateway). Она охватывает методы, описанные в документации [GREEN-API WABA API](https://green-api.com/waba/docs/api/).

## Функциональные возможности

Отправка произвольного текста и файлов внутри 24-часового окна обслуживания клиента. Отправка предварительно одобренных шаблонов сообщений в любое время. Создание, редактирование, получение списка и удаление шаблонов с отслеживанием статуса проверки Meta. Получение статуса и настроек аккаунта, истории чата и журналов. Получение входящих уведомлений через метод длинных опросов (polling) или HTTP-приемник (receiver). Поддержка работы с несколькими инстансами в рамках одного шлюза. Интерактивный виджет шаблонов для хостов, поддерживающих MCP Apps. Метрики Prometheus и интеграция с OpenTelemetry.

## Правила WABA, которые должен знать агент

- **24-часовое окно.** Meta доставляет произвольные сообщения (`waba_send_message`, `waba_send_file`) только внутри 24-часового окна обслуживания, которое открывается последним входящим сообщением клиента. Вне окна единственный вариант — `waba_send_template` с шаблоном в статусе APPROVED.
- **Шаблоны требуют одобрения.** `waba_create_template` отправляет шаблон на проверку в Meta со статусом PENDING. Перед отправкой опрашивайте `waba_get_templates`, пока статус не станет APPROVED. Шаблоны категорий MARKETING и AUTHENTICATION тарифицируются за каждое сообщение; шаблоны UTILITY дешевле.
- **Только личные чаты.** ID чата — номер телефона в международном формате с суффиксом `@c.us`, например `11001234567@c.us`. Группы не поддерживаются.
- **Без QR-кода.** WABA-инстансы подключаются в [console.green-api.com](https://console.green-api.com); шлюз только читает их состояние.

## Архитектура

Чистая Архитектура (Порты и Адаптеры):

```
cmd/server/main.go                    — точка входа, связывание зависимостей (DI)
internal/
  domain/                             — сущности WABA, имена методов, значения шаблонов (ноль зависимостей)
  application/                        — порты (интерфейсы)
  infrastructure/
    auth.go                           — CredentialManager (хранилище учетных данных в оперативной памяти)
    auth_proxy.go, oauth.go           — proxy-авторизация + OAuth 2.0 PKCE для размещённого сервера
    config/config.go                  — конфигурация YAML + env
    greenapi/client.go                — HTTP-клиент WABA
    mcp/
      server.go                       — MCP-сервер (stdio/sse/http/hybrid)
      tools.go, tool_names.go         — регистрация инструментов (tools) MCP и их имена
      tool_meta.go                    — заголовки, подсказки для проверки, привязки виджета
      resources.go                    — ресурсы MCP и виджет шаблонов
      prompts.go, prompt_texts.go     — промпты MCP
    templates/templates_app.html      — виджет шаблонов (MCP App)
    webhook/
      bridge.go                       — мост для длинных опросов (polling)
      receiver.go                     — HTTP-приемник (receiver)
```

## Сборка

```bash
go build -o mcp-gateway-waba ./cmd/server
```

Требуется Go 1.25+.

## Транспорты и эндпоинты

Сервер поддерживает четыре транспорта, выбираемых через параметр `server.transport` (или переменную окружения `GREEN_API_TRANSPORT`):

| Транспорт | Эндпоинты (относительно `host:port`)            | Сценарий использования                                                   |
|-----------|--------------------------------------------------|--------------------------------------------------------------------------|
| `stdio`   | stdin/stdout                                     | Локальные CLI (Claude Desktop, Cursor), запускающие бинарный файл.       |
| `sse`     | `GET /` (stream) + `POST /message`               | Устаревшие MCP-клиенты (только SSE).                                     |
| `http`    | `POST/GET /mcp`                                  | Streamable HTTP (MCP 2025-03-26+) — предпочтительно для новых клиентов.  |
| `hybrid`  | `/sse` + `/message` **и** `/mcp` на одном порту  | Публичный сервер: обслуживает и устаревший SSE, и Streamable HTTP.       |

Все транспорты на базе HTTP также предоставляют:

- `GET /health`, `GET /healthz` — проверки работоспособности (без авторизации)
- `GET /favicon.ico`, `GET /favicon.png` — иконка GREEN-API

Когда включен режим `auth.mode: proxy`, также добавляются эндпоинты OAuth 2.0 PKCE:

- `GET /.well-known/oauth-authorization-server` — метаданные RFC 8414
- `GET /.well-known/oauth-protected-resource` — метаданные RFC 9728
- `GET /authorize` — форма авторизации для браузера
- `POST /token` — обмен токенов

Размещённый публичный сервер работает в режиме `hybrid` по адресу `https://mcp-waba.green-api.com`, поэтому клиенты могут подключаться как к `https://mcp-waba.green-api.com/mcp` (рекомендуется), так и к `https://mcp-waba.green-api.com/sse` (устаревший вариант).

## Конфигурация

### Через YAML

```yaml
server:
  transport: stdio   # stdio | sse | http | hybrid
  port: 8090

auth:
  mode: config       # config | proxy
  cache_ttl: 300     # секунды (только для режима proxy)

green_api:
  instances:
    - id: 1101000001
      api_token: "YOUR_TOKEN"
      api_url: "https://api.green-api.com"

webhook:
  mode: polling
  polling_timeout: 20

rate_limit:
  requests_per_second: 10
  burst: 20

logging:
  level: info
  format: text
```

Все параметры описаны в [config/config.example.yaml](config/config.example.yaml).

### Через переменные окружения

Переменные окружения имеют приоритет над конфигурацией из YAML:

- `GREEN_API_INSTANCE_ID` — ID инстанса
- `GREEN_API_TOKEN` — API-токен
- `GREEN_API_URL` — базовый URL API (по умолчанию `https://api.green-api.com`)
- `GREEN_API_TRANSPORT` — транспорт: `stdio`, `sse`, `http` или `hybrid`
- `GREEN_API_PORT` — HTTP-порт (по умолчанию `8090`)
- `GREEN_API_BASE_URL` — публичный базовый URL (используется для эмиттера/редиректов OAuth, например, `https://mcp.example.com`)
- `GREEN_API_WIDGET_DOMAIN` — домен виджета для метаданных MCP Apps (по умолчанию равен `GREEN_API_BASE_URL`)
- `GREEN_API_AUTH_MODE` — `config` или `proxy`
- `GREEN_API_AUTH_CACHE_TTL` — время жизни (TTL) кэша учетных данных для proxy-авторизации (в секундах)
- `GREEN_API_WEBHOOK_MODE` — `polling` или `receiver`
- `LOG_LEVEL` — уровень логирования
- `LOG_FORMAT` — `text` или `json`

## Запуск

```bash
# С переменными окружения
export GREEN_API_INSTANCE_ID=1101000001
export GREEN_API_TOKEN=your_token
./mcp-gateway-waba

# С файлом конфигурации
./mcp-gateway-waba --config config.yaml
```

## Интеграция

### Claude Desktop (локальный stdio)

Добавьте в `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "waba": {
      "command": "/path/to/mcp-gateway-waba",
      "args": ["--config", "/path/to/config.yaml"]
    }
  }
}
```

### Размещённый сервер (Streamable HTTP + OAuth)

Укажите любому MCP-клиенту с поддержкой Streamable HTTP и OAuth 2.0 адрес размещённого сервера:

```json
{
  "mcpServers": {
    "green-api-waba": {
      "url": "https://mcp-waba.green-api.com/mcp"
    }
  }
}
```

### Docker

```bash
docker build -t mcp-gateway-waba .

docker run --rm -i \
  -e GREEN_API_INSTANCE_ID=1101000001 \
  -e GREEN_API_TOKEN=your_token \
  mcp-gateway-waba
```

## Инструменты MCP

### Сессия

- `waba_connect` — регистрация инстанса в сессии и проверка учетных данных
- `waba_disconnect` — удаление инстанса из сессии

### Отправка

- `waba_send_message` — произвольный текст (24-часовое окно)
- `waba_send_template` — одобренный шаблон с параметрами, медиа-заголовком и postback-кнопками
- `waba_send_file` — файл по URL (24-часовое окно)
- `waba_send_file_by_upload` — файл из base64 через multipart-загрузку

### Аккаунт

- `waba_get_state` — статус аккаунта
- `waba_get_settings` — настройки уведомлений
- `waba_set_settings` — изменение настроек уведомлений
- `waba_get_wa_settings` — номер телефона и статус аккаунта
- `waba_reboot` — перезагрузка инстанса

### Шаблоны

- `waba_create_template` — создание шаблона и отправка на проверку в Meta
- `waba_edit_template` — редактирование шаблона в статусе REJECTED, APPROVED или PAUSED
- `waba_get_templates` — список шаблонов с необязательным фильтром по статусу
- `waba_get_template` — один шаблон по ID
- `waba_delete_template` — удаление по имени элемента
- `waba_delete_template_by_id` — удаление одной версии по ID

### Получение

- `waba_receive_notification` — получение одного уведомления из очереди
- `waba_delete_notification` — подтверждение обработки уведомления

### Журналы

- `waba_get_chat_history` — история чата
- `waba_get_message` — одно сообщение по ID
- `waba_last_incoming_messages` — последние входящие сообщения
- `waba_last_outgoing_messages` — последние исходящие сообщения

### Очереди

- `waba_show_messages_queue` — ожидающие отправки сообщения
- `waba_clear_messages_queue` — очистка очереди исходящих сообщений

### Партнёрские методы

Требуют `partner_token` из партнёрского кабинета GREEN-API, отдельный от учетных данных инстанса.

- `waba_create_instance` — создание нового инстанса
- `waba_delete_instance` — удаление инстанса. Параметры: `partner_token`, `instance_id`
- `waba_get_instances` — список всех инстансов партнёрского аккаунта

### Сервисные методы

- `waba_service_method` — вызов метода GREEN-API по имени. См. [справочник методов](https://green-api.com/docs/api/).

## Ресурсы MCP

- `waba://instance/{id}/state` — статус аккаунта
- `waba://instance/{id}/settings` — настройки уведомлений
- `waba://instance/{id}/templates` — все шаблоны со статусами
- `ui://templates` — интерактивный виджет шаблонов (MCP App), отображается для `waba_get_templates`

## Промпты MCP

- `waba_customer_support` — промпт агента поддержки с правилами 24-часового окна
- `waba_template_broadcast` — планирование и запуск рассылки по шаблону

## Зависимости (Dependencies)

- [mcp-go](https://github.com/mark3labs/mcp-go) — MCP SDK
- [GREEN-API WABA API](https://green-api.com/waba/docs/api/) — справочник HTTP API

## Разработка (Development)

```bash
go mod tidy
go build ./...
go test ./...
gofmt -w .
GOOS=linux GOARCH=amd64 go build -o mcp-gateway-waba-linux ./cmd/server
```

## Лицензия (License)

MIT — см. [LICENSE](LICENSE).
