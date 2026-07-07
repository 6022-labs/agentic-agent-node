# agentic-agent-node

A single-agent runtime — **one agent per container**. Each node gives its agent an
on-chain NFT identity (with an ENS profile), an LLM inference loop that can call
MCP tools, and x402 payments so the agent can charge to speak and settle on-chain.
The runtime DB therefore holds at most one LLM connection and one agent-identity row.

## Quickstart

Prerequisites: **Docker** (with Compose) and, for local Go work, **Go 1.26**.

```bash
docker compose up
```

This runs one container with nginx plus both host binaries, alongside Vault,
Postgres, IPFS, pgAdmin, and a Grafana observability stack. Once it is up:

- **Console:** http://localhost:5500
- **API:** http://localhost:5500/api

nginx listens on port 5000 in-container and is published on host port 5500.

## Architecture at a glance

nginx (in-container port 5000) is the single public listener; nothing else is
reachable from outside. The trust boundary is:

```
external → nginx → gateway_host (wallet auth) → runtime_host (loopback only, no auth)
```

- **`gateway_host`** — the only backend exposed to users: wallet auth
  (challenge nonce → EIP-191 `personal_sign` → JWT), config endpoints, `/responses`;
  it proxies to the runtime over loopback.
- **`runtime_host`** — bound to 127.0.0.1, never exposed: the inference loop,
  knowledge, payment, identity, and blockchain config, plus wallet bootstrap.

The codebase is hexagonal (ports & adapters) with strict tier-service boundaries.
See [`CLAUDE.md`](./CLAUDE.md) for the full architecture, domain-module table, and
the x402 payment model.

## Common commands

A [`Makefile`](./Makefile) wraps the documented workflows:

```bash
make build     # go build ./src/...
make test      # unit tests
make lint      # golangci-lint
make mocks     # regenerate mocks (needs deepsearch-mockgen)
make swagger   # regenerate host API docs
make help      # list targets
```

Without the Makefile, the equivalents are `go build ./src/...` and
`go test $(go list ./tests/... | grep '_unit_tests')`.

## Web console

The React + Vite console lives in [`web/`](./web) and talks only to the gateway
via the relative `/api` base URL. For dev mode and configuration, see
[`web/README.md`](./web/README.md) (`cd web && npm run dev`).

## Layout

Domain modules live under `src/<domain>` (`agent_reasoning`, `agent_identity`,
`agent_knowledge`, `agent_payment`, `runtime_common`, `gateway`), each paired with
its adapter modules. The full module/adapter table is in [`CLAUDE.md`](./CLAUDE.md).
