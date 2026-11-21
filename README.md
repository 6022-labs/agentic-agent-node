# agentic-agent-node

A single-agent runtime — **one agent per container**. Each node gives its agent an
on-chain NFT identity (with an ENS profile), an LLM inference loop that can call
MCP tools, and x402 payments so the agent can charge to speak and settle on-chain.
The runtime DB therefore holds at most one LLM connection and one agent-identity row.

## Quickstart

Prerequisites: **Docker** (with Compose) and, for local Go work, **Go 1.27**.

```bash
docker compose up
```

This runs one container with nginx plus both host binaries, alongside Vault,
Postgres, IPFS, pgAdmin, and a Grafana observability stack. Once it is up:

- **Console:** http://localhost:5500
- **API:** http://localhost:5500/api
- **Grafana:** http://localhost:3000

nginx listens on port 5000 in-container and is published on host port 5500.

Vault runs in dev mode and keeps secrets in memory: restarting it loses the
agent's wallet key while Postgres keeps its data. Reset everything together with
`docker compose down -v`.

### On a local chain

```bash
docker compose -f docker-compose.yml -f docker-compose.local-chain.yml up
```

Adds an Anvil chain (id `31337`, RPC http://localhost:8545) with the 6022 contracts deployed from
[agentic-smart-contracts](https://github.com/6022-labs/agentic-smart-contracts); the agent waits for the deploy.
Names on it don't resolve through public ENS, and it has no x402 facilitator.
The chain state lives in a volume next to Postgres: reset both with `down -v` on
the same two files, never one volume alone.

## Observability

Both hosts export metrics, logs and traces over OTLP (gRPC or HTTP) to Grafana Alloy, which
fans them out to Mimir, Loki and Tempo. Several nodes, apps and environments can
share one stack; `service.name`, `service.instance.id` and
`deployment.environment.name` distinguish them, and every dashboard filters on
them.

Agent side (`TELEMETRY__*`, or the `telemetry` block of `appsettings.json`):

| Setting | Meaning |
| --- | --- |
| `enabled` | export at all |
| `protocol` | `grpc` (default) or `http` — use `http` when the route to Alloy only carries plain HTTPS |
| `endpoint` | Alloy OTLP `host:port` (4317 for gRPC, 4318 for HTTP) |
| `insecure` | disable TLS on the connection |
| `instance_id` | value of `service.instance.id` and `agent.node.id` |
| `environment` | value of `deployment.environment.name` (required when enabled) |
| `auth_token` | bearer credential |
| `auth_username` + `auth_password` | basic credential |

The scheme is derived: a token means bearer, a username and password mean basic,
neither means no authentication. Setting both is fatal at startup.

Alloy side — the OTLP ingest is **unauthenticated by default** and the entrypoint
logs a warning saying so. Set one of these on the alloy container to change that:

| Variable | Effect |
| --- | --- |
| `OTLP_AUTH_HTPASSWD` | basic auth, `user:password` per line — one credential per node, matched against that node's `auth_username` |
| `OTLP_AUTH_TOKEN` | bearer auth with a single token shared by every node |
| `OTLP_TLS_CERT_FILE` + `OTLP_TLS_KEY_FILE` | serve OTLP over TLS |
| `OTLP_TLS_CLIENT_CA_FILE` | additionally require a client certificate (mTLS) |

`OTLP_AUTH_HTPASSWD` wins when both are set. Its content is read with go-htpasswd,
so bcrypt/md5/sha hashes and plaintext lines all work — prefer hashes, and escape
`$` as `$$` inside `docker-compose.yml`.

Authentication over a plaintext connection sends the credential in the clear: pair
it with TLS on any untrusted network. The agent logs a warning when it is asked to
do otherwise.

## Database TLS

The runtime database is set with `RUNTIME_DATABASE__*` (or the `runtime_database`
block of `appsettings.json`). By default TLS is whatever the `url` asks for. To
reach a database whose certificate is signed by a private CA, for example through
a TCP proxy, add:

| Setting | Meaning |
| --- | --- |
| `ca_certificate` | PEM of the CA the server certificate must chain to |
| `tls_server_name` | name the certificate is checked against; defaults to the `url` host |

With `ca_certificate` set the connection always uses TLS 1.2+ with full
verification, whatever the `url` says (`sslmode`, `tls=`): there is no plaintext
or skip-verify fallback. It works with `postgres`, `mysql` and `mariadb`; with
`sqlite`, an unparsable PEM, or `tls_server_name` alone, startup is fatal. On
Postgres only the first host of the `url` is used.

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
make hooks     # git hooks: format + secret scan on commit, build/lint/test on push
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

## License

MIT — see [LICENSE](LICENSE).
