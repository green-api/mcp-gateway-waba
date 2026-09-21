# GREEN-API WABA Skill

OpenClaw skill for messaging customers from an official WhatsApp Business API (WABA) number via [GREEN-API](https://green-api.com/en/waba).

## What it does

Gives OpenClaw access to 29 tools: template and session messaging, file sending, template management with Meta review status, account settings, notifications, journals, queues, and instance provisioning via the Partner API — all through the GREEN-API WABA MCP gateway.

## Prerequisites

1. A GREEN-API account — sign up at [console.green-api.com](https://console.green-api.com)
2. A connected WABA instance (state `authorized` in the console)
3. The GREEN-API WABA MCP gateway server running (see below)

## Setup

### 1. Install the skill

```bash
openclaw skills install green-api-waba
```

### 2. Connect the MCP server

Add to your OpenClaw config (`~/.openclaw/openclaw.json`):

```json
{
  "mcpServers": {
    "green-api-waba": {
      "url": "https://mcp-waba.green-api.com/mcp"
    }
  }
}
```

The hosted server runs in hybrid mode, so legacy SSE-only clients can use `https://mcp-waba.green-api.com/sse` instead.

### 3. Authorize

On the first request, the MCP client will open a browser tab to `/authorize`. Paste your **Instance ID** and **API Token** (from [console.green-api.com](https://console.green-api.com)) into the form. The bearer token issued by the OAuth flow is bound to those credentials for 24 hours.

### 4. Start messaging

Once authorized, just describe what you want:

> "Send the order_update template to +1 100 123 4567 with order number 42"

## Links

- [GREEN-API WABA Documentation](https://green-api.com/en/waba/docs/api/)
- [GREEN-API Console](https://console.green-api.com)
