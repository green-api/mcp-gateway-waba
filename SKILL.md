---
name: green-api-waba
version: 1.0.0
description: Send WhatsApp Business (WABA) template and session messages, manage message templates and receive notifications via the GREEN-API WABA MCP gateway.
homepage: https://green-api.com/en/waba
mcp_endpoint: https://mcp-waba.green-api.com/mcp
metadata: { "openclaw": { "emoji": "✅" } }
---

# GREEN-API WABA Skill

Use this skill when the user wants to message customers from an **official WhatsApp Business API (WABA)** number through GREEN-API: send approved templates, reply inside the 24-hour window, manage templates or read notifications.

## Connection

MCP endpoint (Streamable HTTP, preferred): `https://mcp-waba.green-api.com/mcp`
MCP endpoint (legacy SSE): `https://mcp-waba.green-api.com/sse`

The hosted endpoint uses OAuth 2.0 (Authorization Code + PKCE): the user authenticates once in a browser, after which the bearer token is bound to their `instance_id` + `api_token`. In this mode `instance_id` can be omitted from tool calls and `waba_connect` is not required.

For self-hosted or stdio deployments call `waba_connect` first with `instance_id` and `api_token` (from console.green-api.com). Call `waba_disconnect` when done.

## Rules that differ from regular WhatsApp

- **24-hour window.** `waba_send_message` and `waba_send_file` reach the customer only inside the 24-hour customer service window opened by the customer's last incoming message. Check `waba_last_incoming_messages` first. Outside the window use `waba_send_template`.
- **Templates must be APPROVED.** `waba_create_template` submits a template to Meta (status PENDING). Only APPROVED templates can be sent. MARKETING and AUTHENTICATION templates are billed per message, UTILITY templates are cheaper.
- **Personal chats only.** Chat ID = phone number in international format + `@c.us`, for example `11001234567@c.us`. There are no groups, polls, contacts, locations or message edits.
- **No QR code.** WABA instances are connected in console.green-api.com. `waba_get_state` reports `authorized`, `notAuthorized`, `blocked`, `sleepMode`, `starting` or `yellowCard`.

## Tools

### Session

| Tool | Description |
|---|---|
| `waba_connect` | Register an instance (self-hosted only). Params: `instance_id`, `api_token`, optional `api_url` |
| `waba_disconnect` | Remove the instance from the session |

### Sending

| Tool | Description |
|---|---|
| `waba_send_message` | Free-form text inside the 24-hour window. Params: `instance_id`, `chat_id`, `message`, optional `link_preview` |
| `waba_send_template` | Send an APPROVED template. Params: `instance_id`, `chat_id`, `template_id`, `params` (values for `{{1}}`, `{{2}}`…), optional `media_type` + `media_url` (+ `media_filename` for documents), `message` (raw object for LOCATION/CAROUSEL), `postback_texts` |
| `waba_send_file` | File by URL inside the window. Params: `instance_id`, `chat_id`, `url`, `file_name`, optional `caption` |
| `waba_send_file_by_upload` | File from base64. Params: `instance_id`, `chat_id`, `file_base64`, `filename`, optional `caption` |

### Account

| Tool | Description |
|---|---|
| `waba_get_state` | Account state |
| `waba_get_settings` | Notification settings |
| `waba_set_settings` | Update settings. Params: optional `webhook_url` (empty string disables), `webhook_url_token`, `outgoing_webhook`, `outgoing_api_message_webhook`, `incoming_webhook` |
| `waba_get_wa_settings` | Account phone number and state |
| `waba_reboot` | Reboot the instance |

### Templates

| Tool | Description |
|---|---|
| `waba_create_template` | Create and submit a template. Params: `element_name`, `language_code`, `category` (AUTHENTICATION/MARKETING/UTILITY), `template_type` (TEXT/IMAGE/VIDEO/DOCUMENT/CAROUSEL), `vertical`, `content`, `example`, optional `header`, `example_header`, `footer`, `buttons`, `cards`, `media_url`, `enable_sample`, `add_security_recommendation`, `code_expiration_minutes` |
| `waba_edit_template` | Edit a REJECTED, APPROVED or PAUSED template. Params: `template_id` plus the fields to change |
| `waba_get_templates` | List templates with status. Optional `status` filter |
| `waba_get_template` | One template. Params: `template_id` |
| `waba_delete_template` | Delete by name. Params: `element_name` |
| `waba_delete_template_by_id` | Delete one version. Params: `element_name`, `template_id` |

### Receiving

| Tool | Description |
|---|---|
| `waba_receive_notification` | Take one notification from the queue (incoming messages, template button replies, statuses) |
| `waba_delete_notification` | Acknowledge it. Params: `instance_id`, `receipt_id` |

### Journals and queues

| Tool | Description |
|---|---|
| `waba_get_chat_history` | Params: `instance_id`, `chat_id`, optional `count` (default 50, max 100) |
| `waba_get_message` | Params: `instance_id`, `chat_id`, `id_message` |
| `waba_last_incoming_messages` | Params: `instance_id`, optional `minutes` (default 60, max 1440) |
| `waba_last_outgoing_messages` | Params: `instance_id`, optional `minutes` (default 60, max 1440) |
| `waba_show_messages_queue` | Pending outgoing messages |
| `waba_clear_messages_queue` | Drop the outgoing queue |

### Partner API

Require a `partner_token` from the GREEN-API partner dashboard, separate from any instance's own credentials.

| Tool | Description |
|---|---|
| `waba_create_instance` | Create a new instance. Params: `partner_token` |
| `waba_delete_instance` | Delete an instance. Params: `partner_token`, `instance_id` |
| `waba_get_instances` | List all instances under the partner account. Params: `partner_token` |

### Service

| Tool | Description |
|---|---|
| `waba_service_method` | Call a GREEN-API method by name. See [the method reference](https://green-api.com/en/docs/api/). Params: `instance_id`, `request` |

## Resources

- `waba://instance/{id}/state` — account state
- `waba://instance/{id}/settings` — notification settings
- `waba://instance/{id}/templates` — templates with status
- `ui://templates` — interactive templates widget (MCP App)

## Prompts

- `waba_customer_support` — support agent prompt. Args: `instance_id`, `language`
- `waba_template_broadcast` — template broadcast plan. Args: `instance_id`, `template_id`
