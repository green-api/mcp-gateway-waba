# mcp-gateway-waba

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/green-api/mcp-gateway-waba/main)](https://go.dev/)
[![release](https://img.shields.io/github/v/release/green-api/mcp-gateway-waba)](https://github.com/green-api/mcp-gateway-waba/releases)
[![Protocol: MCP](https://img.shields.io/badge/Protocol-MCP-orange.svg)](https://modelcontextprotocol.io)
[![Go Report Card](https://goreportcard.com/badge/github.com/green-api/mcp-gateway-waba)](https://goreportcard.com/report/github.com/green-api/mcp-gateway-waba)

- [Документация на русском](README_RU.md)

## Support Links

[![Support](https://img.shields.io/badge/support@green--api.com-D14836?style=for-the-badge&logo=gmail&logoColor=white)](mailto:support@green-api.com)
[![Support](https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white)](https://t.me/greenapi_support_eng_bot)
[![Support](https://img.shields.io/badge/WhatsApp-25D366?style=for-the-badge&logo=whatsapp&logoColor=white)](https://wa.me/77780739095)

## Guides & News

[![Guides](https://img.shields.io/badge/YouTube-%23FF0000.svg?style=for-the-badge&logo=YouTube&logoColor=white)](https://www.youtube.com/@greenapi-en)
[![News](https://img.shields.io/badge/Telegram-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white)](https://t.me/green_api)
[![News](https://img.shields.io/badge/WhatsApp-25D366?style=for-the-badge&logo=whatsapp&logoColor=white)](https://whatsapp.com/channel/0029VaLj6J4LNSa2B5Jx6s3h)

---

MCP gateway for GREEN-API **WABA**, the official WhatsApp Business API. Lets AI agents (Claude Desktop, ChatGPT, Cursor, OpenClaw and others) send template and session messages, manage message templates and receive notifications from a WhatsApp Business account through the [Model Context Protocol](https://modelcontextprotocol.io).

This is the WABA counterpart of [green-api-mcp-gateway](https://github.com/green-api/green-api-mcp-gateway). It covers the methods documented in the [GREEN-API WABA API](https://green-api.com/en/waba/docs/api/).

## What it does

Send free-form text and files inside the 24-hour customer service window. Send pre-approved message templates at any time. Create, edit, list and delete templates and track their Meta review status. Read account state and settings, chat history and journals. Receive incoming notifications via polling or an HTTP receiver. Multi-instance support in a single gateway. An interactive templates widget for hosts that render MCP Apps. Prometheus metrics and OpenTelemetry.

## WABA rules the agent must know

- **24-hour window.** Meta delivers free-form messages (`waba_send_message`, `waba_send_file`) only inside the 24-hour customer service window opened by the customer's last incoming message. Outside it the only option is `waba_send_template` with an APPROVED template.
- **Templates need approval.** `waba_create_template` submits a template to Meta with status PENDING. Poll `waba_get_templates` until it is APPROVED before sending it. MARKETING and AUTHENTICATION templates are billed per message; UTILITY templates are cheaper.
- **Personal chats only.** Chat IDs are phone numbers in international format followed by `@c.us`, for example `11001234567@c.us`. Groups are not supported.
- **No QR code.** WABA instances are connected in [console.green-api.com](https://console.green-api.com); the gateway only reads their state.

## Architecture

Clean Architecture (Ports & Adapters):

```
cmd/server/main.go                    — entry point, DI wiring
internal/
  domain/                             — WABA entities, method names, template values (zero dependencies)
  application/                        — ports (interfaces)
  infrastructure/
    auth.go                           — CredentialManager (in-memory credential store)
    auth_proxy.go, oauth.go           — proxy auth + OAuth 2.0 PKCE for the hosted server
    config/config.go                  — YAML + env configuration
    greenapi/client.go                — WABA HTTP client
    mcp/
      server.go                       — MCP server (stdio/sse/http/hybrid)
      tools.go, tool_names.go         — MCP tools registration and names
      tool_meta.go                    — titles, review hints, widget bindings
      resources.go                    — MCP resources and the templates widget
      prompts.go, prompt_texts.go     — MCP prompts
    templates/templates_app.html      — templates widget (MCP App)
    webhook/
      bridge.go                       — polling bridge
      receiver.go                     — HTTP receiver
```

## Build

```bash
go build -o mcp-gateway-waba ./cmd/server
```

Requires Go 1.25+.

## Transports & endpoints

The server supports four transports, selected via `server.transport` (or the `GREEN_API_TRANSPORT` env var):

| Transport | Endpoints (relative to `host:port`)             | Use case                                                   |
|-----------|--------------------------------------------------|------------------------------------------------------------|
| `stdio`   | stdin/stdout                                     | Local CLIs (Claude Desktop, Cursor) launching the binary.  |
| `sse`     | `GET /` (stream) + `POST /message`               | Legacy MCP clients (SSE only).                             |
| `http`    | `POST/GET /mcp`                                  | Streamable HTTP (MCP 2025-03-26+) — preferred for new clients. |
| `hybrid`  | `/sse` + `/message` **and** `/mcp` on one port   | Public server: serves both legacy SSE and Streamable HTTP. |

All HTTP-based transports also expose:

- `GET /health`, `GET /healthz` — health probes (no auth)
- `GET /favicon.ico`, `GET /favicon.png` — GREEN-API icon

When `auth.mode: proxy` is enabled, the OAuth 2.0 PKCE endpoints are added too:

- `GET /.well-known/oauth-authorization-server` — RFC 8414 metadata
- `GET /.well-known/oauth-protected-resource` — RFC 9728 metadata
- `GET /authorize` — browser-facing authorization form
- `POST /token` — token exchange

The hosted public deployment runs in `hybrid` mode at `https://mcp-waba.green-api.com`, so clients can connect to either `https://mcp-waba.green-api.com/mcp` (recommended) or `https://mcp-waba.green-api.com/sse` (legacy).

## Configuration

### Via YAML

```yaml
server:
  transport: stdio   # stdio | sse | http | hybrid
  port: 8090

auth:
  mode: config       # config | proxy
  cache_ttl: 300     # seconds (proxy mode only)

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

See [config/config.example.yaml](config/config.example.yaml) for every option.

### Via environment variables

Environment variables take precedence over the YAML config:

- `GREEN_API_INSTANCE_ID` — instance ID
- `GREEN_API_TOKEN` — API token
- `GREEN_API_URL` — API base URL (defaults to `https://api.green-api.com`)
- `GREEN_API_TRANSPORT` — transport: `stdio`, `sse`, `http`, or `hybrid`
- `GREEN_API_PORT` — HTTP port (defaults to `8090`)
- `GREEN_API_BASE_URL` — public base URL (used for OAuth issuer/redirects, e.g. `https://mcp.example.com`)
- `GREEN_API_WIDGET_DOMAIN` — widget origin for MCP Apps submission metadata (defaults to `GREEN_API_BASE_URL`)
- `GREEN_API_AUTH_MODE` — `config` or `proxy`
- `GREEN_API_AUTH_CACHE_TTL` — proxy-auth credential cache TTL (seconds)
- `GREEN_API_WEBHOOK_MODE` — `polling` or `receiver`
- `LOG_LEVEL` — log level
- `LOG_FORMAT` — `text` or `json`

## Run

```bash
# With environment variables
export GREEN_API_INSTANCE_ID=1101000001
export GREEN_API_TOKEN=your_token
./mcp-gateway-waba

# With a config file
./mcp-gateway-waba --config config.yaml
```

## Integration

### Claude Desktop (local stdio)

Add to `claude_desktop_config.json`:

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

### Hosted server (Streamable HTTP + OAuth)

Point any MCP client that supports Streamable HTTP and OAuth 2.0 at the hosted endpoint:

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

## MCP Tools

### Session

- `waba_connect` — register an instance for the session and validate credentials
- `waba_disconnect` — remove an instance from the session

### Sending

- `waba_send_message` — free-form text (24-hour window)
- `waba_send_template` — pre-approved template with parameters, media header, postbacks
- `waba_send_file` — file by URL (24-hour window)
- `waba_send_file_by_upload` — file from base64 via multipart upload

### Account

- `waba_get_state` — account state
- `waba_get_settings` — notification settings
- `waba_set_settings` — update notification settings
- `waba_get_wa_settings` — account phone and state
- `waba_reboot` — reboot the instance

### Templates

- `waba_create_template` — create and submit a template to Meta
- `waba_edit_template` — edit a REJECTED, APPROVED or PAUSED template
- `waba_get_templates` — list templates, optional status filter
- `waba_get_template` — one template by ID
- `waba_delete_template` — delete by element name
- `waba_delete_template_by_id` — delete one version by ID

### Receiving

- `waba_receive_notification` — take one notification from the queue
- `waba_delete_notification` — acknowledge a notification

### Journals

- `waba_get_chat_history` — chat history
- `waba_get_message` — one message by ID
- `waba_last_incoming_messages` — recent incoming messages
- `waba_last_outgoing_messages` — recent outgoing messages

### Queues

- `waba_show_messages_queue` — pending outgoing messages
- `waba_clear_messages_queue` — drop the outgoing queue

### Partner API

These require a `partner_token` from the GREEN-API partner dashboard, separate from any instance's own credentials.

- `waba_create_instance` — create a new instance
- `waba_delete_instance` — delete an instance. Params: `partner_token`, `instance_id`
- `waba_get_instances` — list all instances under the partner account

### Service

- `waba_service_method` — call a GREEN-API method by name. See the [method reference](https://green-api.com/en/docs/api/).

## MCP Resources

- `waba://instance/{id}/state` — account state
- `waba://instance/{id}/settings` — notification settings
- `waba://instance/{id}/templates` — all templates with status
- `ui://templates` — interactive templates widget (MCP App), rendered for `waba_get_templates`

## MCP Prompts

- `waba_customer_support` — support agent prompt with the 24-hour window rules
- `waba_template_broadcast` — plan and run a template broadcast

## Dependencies

- [mcp-go](https://github.com/mark3labs/mcp-go) — MCP SDK
- [GREEN-API WABA API](https://green-api.com/en/waba/docs/api/) — HTTP API reference

## Development

```bash
go mod tidy
go build ./...
go test ./...
gofmt -w .
GOOS=linux GOARCH=amd64 go build -o mcp-gateway-waba-linux ./cmd/server
```

## License

MIT — see [LICENSE](LICENSE).
