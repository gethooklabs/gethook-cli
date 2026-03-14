# GetHook CLI

> Receive, debug, and replay webhooks locally — in seconds.

<!--
  Replace this comment with an animated GIF showing:
  $ gethook listen --forward-to http://localhost:3000/webhooks
  ╭──────────────────────────────────────────────────────╮
  │ ✓ Tunnel active                                       │
  │   https://in.gethook.dev/src_d7f3a1                  │
  │ → Forwarding to http://localhost:3000/webhooks        │
  │   Press Ctrl+C to stop.                              │
  ╰──────────────────────────────────────────────────────╯
  14:03:01  POST  stripe.payment_intent.succeeded  200 OK  (45ms)
  14:03:12  POST  stripe.charge.failed             502      retrying in 30s
-->

**GetHook CLI** is the fastest way to work with webhooks locally. No more public tunnels,
no more digging through dashboards — just one command and you're receiving webhooks.

[![GitHub release](https://img.shields.io/github/v/release/gethook/gethook-cli?style=flat-square)](https://github.com/gethook/gethook-cli/releases)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue?style=flat-square)](LICENSE)
[![CI](https://github.com/gethook/gethook-cli/actions/workflows/test.yml/badge.svg)](https://github.com/gethook/gethook-cli/actions)

---

## Why

| Problem | GetHook CLI |
|---|---|
| Webhooks can't reach `localhost` | Live tunnel with one command |
| Debugging failed deliveries means opening dashboards | `gethook inspect evt_123` in your terminal |
| Replaying events requires manual API calls | `gethook replay evt_123 --forward-to localhost:3000` |
| Testing Stripe webhooks without Stripe | `gethook trigger stripe payment_intent.succeeded` |

---

## Install

**macOS / Linux (Homebrew):**
```bash
brew install gethook/tap/gethook
```

**macOS / Linux (curl):**
```bash
curl -fsSL https://cli.gethook.dev/install.sh | sh
```

**npm (for Node.js projects):**
```bash
npm install -g @gethook/cli
```

**Windows (Scoop):**
```bash
scoop bucket add gethook https://github.com/gethook/scoop-bucket
scoop install gethook
```

**Go:**
```bash
go install github.com/gethook/gethook-cli@latest
```

Or download a binary directly from [GitHub Releases](https://github.com/gethook/gethook-cli/releases).

---

## 60-second quickstart

```bash
# 1. Log in (or skip for anonymous mode)
gethook login

# 2. Start listening
gethook listen --forward-to http://localhost:3000/webhooks

# Output:
# ╭──────────────────────────────────────────────────────────╮
# │ ✓ Tunnel active                                           │
# │   https://in.gethook.dev/src_abc123                      │
# │ → Forwarding to http://localhost:3000/webhooks            │
# │   Paste the tunnel URL into your webhook provider.        │
# │   Press Ctrl+C to stop.                                   │
# ╰──────────────────────────────────────────────────────────╯
#
# 14:03:01  POST  stripe.payment_intent.succeeded  200 OK  (45ms)

# 3. Paste the tunnel URL into Stripe / GitHub / Shopify
# 4. Trigger an action — watch it arrive in your terminal
```

**That's it.** Webhooks are flowing to your local server.

---

## Commands

### `gethook listen` — Live webhook tunnel

```bash
# Forward all events to a local server
gethook listen --forward-to http://localhost:3000/webhooks

# Filter to specific event types (glob)
gethook listen --forward-to http://localhost:3000/webhooks --filter "stripe.*"

# Listen on an existing source
gethook listen --source src_abc123

# Print-only mode (no forwarding)
gethook listen --no-tunnel
```

### `gethook events` — View event history

```bash
# List recent events
gethook events

# Stream new events in real time
gethook events --tail

# Filter by status
gethook events --status dead_letter --limit 10

# Pipe-friendly JSON output
gethook events --json | jq '.[] | .event_type'
```

### `gethook replay` — Re-run a past event

```bash
# Replay to original destination
gethook replay evt_abc123

# Replay to your local server (great after fixing a bug)
gethook replay evt_abc123 --forward-to http://localhost:3000/webhooks

# Inspect what would be sent without sending
gethook replay evt_abc123 --dry-run
```

### `gethook inspect` — Full event details

```bash
# Show payload, headers, and all delivery attempts
gethook inspect evt_abc123
```

```
  Event  evt_abc123
  Type   stripe.payment_intent.succeeded
  Status dead_letter (5 attempts)
  Time   2024-01-15 14:03:01 UTC

── Payload ──────────────────────────────────────────────────────
{
  "id": "pi_abc",
  "amount": 2000,
  ...
}

── Delivery Attempts ────────────────────────────────────────────
  #1  14:03:01  timeout      (30s)
  #2  14:03:31  http_5xx     503
  #3  14:05:31  success      200 ✓
```

### `gethook trigger` — Send test payloads

No real API account needed. Payloads are bundled in the CLI.

```bash
# List all available providers and event types
gethook trigger --list

# Send a realistic Stripe payment
gethook trigger stripe payment_intent.succeeded --forward-to http://localhost:3000/webhooks

# Override specific fields
gethook trigger stripe charge.failed --data '{"amount": 0}' --forward-to http://localhost:3000/webhooks

# Available providers: stripe, github, shopify, slack
gethook trigger github push --forward-to http://localhost:3000/webhooks
gethook trigger shopify order.created --forward-to http://localhost:3000/webhooks
```

### `gethook sources` / `gethook destinations` — Manage infrastructure

```bash
# Sources
gethook sources list
gethook sources create my-source
gethook sources delete src_abc123

# Destinations
gethook destinations list
gethook destinations create my-dest https://my-api.example.com/webhooks
gethook destinations delete dest_abc123
```

---

## Workflow examples

### Debug a failing webhook locally

```bash
# 1. Find dead-letter events
gethook events --status dead_letter

# 2. Inspect the payload and failure reason
gethook inspect evt_abc123

# 3. Fix your handler code, restart local server

# 4. Replay the exact same real payload
gethook replay evt_abc123 --forward-to http://localhost:3000/webhooks
# ← 200 OK (38ms)
```

### Test handler without a real Stripe account

```bash
gethook listen --forward-to http://localhost:3000/webhooks &

gethook trigger stripe payment_intent.succeeded
gethook trigger stripe charge.failed
gethook trigger stripe checkout.session.completed --data '{"amount_total": 4900}'
```

### CI webhook integration tests

```bash
# Start your app in the background
npm run dev &

# Fire realistic payloads and assert 200
gethook trigger stripe payment_intent.succeeded --forward-to http://localhost:3000/webhooks
```

---

## How it works

```
Provider (Stripe, GitHub, etc.)
       │
       │  POST to tunnel URL
       ▼
┌──────────────────┐       SSE / polling      ┌─────────────┐
│  GetHook Cloud   │ ─────────────────────►   │ gethook CLI │
│  (relay)         │                          │             │
└──────────────────┘                          │  prints +   │
                                              │  forwards   │
                                              └──────┬──────┘
                                                     │
                                                     │  POST
                                                     ▼
                                           http://localhost:3000
```

The CLI creates a source on your GetHook account, giving you a stable public URL
(`https://in.gethook.dev/src_*`). Inbound webhooks are stored in GetHook's cloud
and streamed to the CLI, which forwards them to your local server.

---

## Configuration

Config is stored at `~/.config/gethook/config.toml`.

| Environment variable | Description |
|---|---|
| `GETHOOK_API_KEY` | Override stored API key |
| `GETHOOK_API_BASE` | Override API base URL (for self-hosted) |
| `GETHOOK_INGEST_BASE` | Override ingest base URL |

For local development against a self-hosted GetHook instance:

```bash
GETHOOK_API_BASE=http://localhost:8080 GETHOOK_INGEST_BASE=http://localhost:8080/ingest gethook listen
```

---

## Full platform

The CLI is the developer-facing entry point. For production webhook infrastructure —
retries, routing, white-labeling, event history, custom domains — see the
[GetHook platform](https://gethook.dev).

---

## Contributing

Issues and PRs are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

```bash
git clone https://github.com/gethook/gethook-cli
cd gethook-cli
go build -o gethook .
./gethook --help
```

To test against a local GetHook backend:

```bash
make run-local ARGS="listen --forward-to http://localhost:3001/webhooks"
```

---

## License

Apache 2.0 — see [LICENSE](LICENSE).
