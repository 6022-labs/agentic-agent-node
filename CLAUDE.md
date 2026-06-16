# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Code style

### Comments — be sharp, write fewer

You write too many comments. Cut hard. The defaults:

- **One line, max.** A comment is one line. The *only* exception is an interface or interface method, which may use **two**. Never write a multi-line or multi-paragraph doc block on anything.
- **Only say the non-obvious.** State a *why*, an invariant, or a cross-package contract. Never restate what the code already says. If the code is clear, write **no** comment — that is the common case.
- **Go doc form.** Doc comments start with the identifier name, then stop (`AgentMinter mints ...`).
- **Document the contract once, on the port.** Method doc comments live on the `*Interface`; the implementation's methods carry **no** doc comment (it only drifts). The impl keeps its type-level comment and comments on private helpers. See `RegistrationReconciler` / `AgentCollectionRequester`.

### Naming

- **Name dependency fields/params after their full type** (the mvc-controller convention), not a generic role. Drop the package qualifier and `Interface` suffix, camelCase the rest: `EnsPublisherInterface` → `ensPublisher` (not `publisher`), `AgentRegistrationStateRepositoryInterface` → `agentRegistrationStateRepository` (not `stateRepository`), `*settings.AgentIdentitySettings` → `agentIdentitySettings`. Idiomatic infra names (`logger`, `client`, `ctx`) and local *domain-value* variables (`agent`, `state`, `collection`, `tokenId`) keep their short names.

## Architecture

Single Go module (`github.com/6022-labs/agentic-agent-node`, Go 1.26), **two binaries** + the web console, all in one image. **One agent = one container** — so the runtime DB holds at most **one** LLM connection row and **one** agent-identity row.

### Hexagonal (ports & adapters) — enforced by structure, not lint

Each **domain module** is a directory under `src/` with `configurations/ services/ services/interfaces/ repositories/ models/ settings/ use_cases/`. The **ports** are the `interfaces/` packages (service ports under `services/interfaces/`, persistence ports under `repositories/`). The **adapters** are *separate sibling modules* that implement those ports (`*_blockchain`, `*_http_*`, `*_mvc`, `runtime_db`, `runtime_vault`).

Rules that the layout exists to enforce — keep them:

- **Domain logic lives in the domain module only.** Adapters translate (HTTP/RPC/SQL/Vault ↔ domain types) and do nothing else. No business rules, no orchestration, no validation-with-meaning in an adapter.
- **Controllers are transport only.** A `*_mvc` controller decodes the request, calls exactly one use case, and maps the result to a status via `common/mvc.WriteError`. **No business logic or business rules in controllers** — if you find a decision being made in a controller, it belongs in a use case. Don't return fiber/HTTP types from use cases.
- **Use cases are the entry points to the domain.** They return idiomatic `(*response, error)` (or just `error` for no-body ops) and depend only on ports.
- Adding a behavior usually means: define/extend a port in `services/interfaces/`, implement the domain logic in `services/` or `use_cases/`, implement the port in the relevant adapter module, wire it in DI.

### Public exposure — the gateway is the only externally reachable backend

nginx (port **5000** in-container) is the single public listener. **Nothing else is reachable from outside.**

- `/` → web console (SPA from `web/dist`).
- `/api/*` → proxied to **`gateway_host`** on loopback, **`/api` prefix stripped** (gateway routes stay at root).
- `/.well-known/*` → proxied verbatim to `gateway_host` (reserved for the A2A agent card).

- **`gateway_host`** (loopback, behind nginx) — the only backend exposed to end users. Wallet auth (challenge nonce → EIP-191 `personal_sign` → JWT), config endpoints, `/responses`. After auth it proxies to the runtime over loopback via `gateway_http_runtime`.
- **`runtime_host`** (bound to 127.0.0.1, internal only — **never** exposed) — everything under `/internal`: inference loop, knowledge, payment, identity, blockchain config. Owns wallet bootstrap, which runs **synchronously at startup** before it serves.
- **`web/`** (React + Vite) — built in the image's `web-builder` stage, served by nginx; talks to the gateway via the relative `/api` base URL.

So the trust boundary is: external → nginx → `gateway_host` (auth) → `runtime_host` (no auth, trusts loopback only). Never add an externally-routable path to the runtime.

### Domain modules and their adapters

| Domain (`src/<domain>`) | Purpose | Adapter modules implementing its ports |
|---|---|---|
| `agent_reasoning` | inference loop, LLM + MCP connections | `agent_reasoning_http_inference` (per-provider LLM clients: anthropic/openai/bedrock/ollama/…), `agent_reasoning_http_mcp` (MCP tool calls), `agent_reasoning_mvc` (controllers), `runtime_db` (connection rows), `runtime_vault` (api keys, MCP header-auth) |
| `agent_identity` | on-chain NFT identity, ENS, wallet bootstrap | `agent_identity_blockchain` (ERC721 mint via AgentCollectionsManager, ENS text records via AgentEnsRegistry, $6022 token reads; shared RPC/transactor/signing infra; all contract bindings in `abi`), `agent_identity_http_ipfs` (image/metadata to IPFS), `agent_identity_mvc`, `runtime_db`, `runtime_vault` (wallet key) |
| `agent_knowledge` | conversation transcripts | `agent_knowledge_mvc`, `runtime_db` |
| `agent_payment` | payment policy + rules, x402 protocol | `agent_payment_blockchain` (on-chain ERC-20 validation), `agent_payment_mvc`, `runtime_db`; settled via the x402 facilitator (see below) |
| `runtime_common` | shared blockchain/operator config (chains, $6022 token, payment facilitators, collection managers, ENS registries) read from settings | `runtime_common_mvc` (exposes blockchain config) |
| `gateway` | auth + config proxy | `gateway_http_runtime` (HTTP client to runtime), `gateway_http_x402` (x402 facilitator client), `gateway_mvc` (fiber controllers) |

### Payment (x402)

`agent_payment` is method-agnostic via a strategy pattern (`PaymentMethodStrategiesInterface` dispatches by `Method`); **x402** is the implemented method. An x402 rule (`PaymentRuleX402Config`) carries a CAIP-2 `network`, ERC-20 `asset` (default $6022), `payTo`, and `assetTransferMethod` (ERC-3009 `transferWithAuthorization` or permit2). **Zero-price rules never settle on-chain** — they carry only EIP-712 signing context. Actual settlement goes through an **x402 facilitator** (`gateway_http_x402`, long timeout because `/settle` blocks until the tx lands); facilitators are configured in `runtime_common/settings/payment_facilitators_settings.go` and a facilitator is only valid if a matching chain RPC exists in `blockchain.chains`.

### Identity resolves lazily

`AgentIdentity` is absent until an NFT is minted; `ErrIdentityNotFound` is the shared sentinel returned by both the in-process resolver and the gateway HTTP client. `RegistrationReconciler` loads confirmed on-chain facts (owner, creator, token) and persists them into the agent + registration-state rows (used by check-mint and self-register).

## Cross-cutting

- **DI**: `go.uber.org/dig`. Each module exposes `Add<Module>Configuration(container)`; all are wired in the host's `configurations/di_configuration.go`. Add a module by adding its `Add...` call there. Controllers register into dig groups (`gateway_controllers` / `runtime_controllers`) and are injected as slices for route registration.
- **Config — every setting must be configurable from BOTH `appsettings` JSON AND env/`.env`.** `koanf` loads `appsettings*.json` then overlays env vars; nesting uses double-underscore lowercased to dotted keys (`JWT__SIGNING_KEY` → `jwt.signing_key`, `GATEWAY_HOST__API_PORT` → `gateway_host.api_port`). Settings structs live in each module's `settings/` with `koanf:"..."` tags and sane defaults in their `New...` constructor. When you add a setting, give it a tag, a default, and confirm the env override path works. Examples: `gateway_host/appsettings.example.json`, `runtime_host/appsettings.example.json`. (See the `koanf-settings` skill.)
- **Secrets never touch the DB or config rows** — `runtime_vault` (HashiCorp Vault, KV v1) holds LLM api keys, the agent wallet credentials, and MCP header-auth values. Rows store only the *key/header names*.
- **Errors → HTTP**: use cases fail with concrete types from `common/errors`, mapped by `common/mvc.WriteError`: `ValidationError`→400, `UnauthorizedError`→401, `NotFoundError`→404, `ConflictError`→409, `UnavailableError`→502, `InternalError`→500; a plain error → 500. Controllers pick the success status (200/201/204).
- **Persistence**: `runtime_db` is gorm; migrations are generated per driver (postgres/mysql/mariadb) via atlas+gorm.

## Commands

```bash
go build ./src/...                                            # build everything (CI: go-build.yml)
find . -type d -name '*unit_tests' -exec go test {}/... \;   # all unit tests (CI: go-unit-tests.yml)
go test ./tests/<package>_unit_tests/...                     # one package
```

- **Integration tests** (`tests/*_integration_tests`) spin up real Postgres/Vault via `tests/common_tests/container_builders` — they need Docker and are **not** in the unit-test CI sweep.
- **Mocks** are generated, never hand-written. Regenerate after changing any interface: `deepsearch-mockgen -S ./src -O ./tests -A -P` (CLI: https://github.com/FournyP/deepsearch-mockgen-cli).
- **Migrations**: `src/runtime_db/migrations/generate_migrations.sh` (atlas+gorm; per-driver SQL under `postgres/`, `mysql/`, `mariadb/`).
- **Swagger docs**: `./scripts/generate_swagger_docs_for_hosts.sh [--host gateway_host]` — regenerates `*_host/docs/{docs.go,swagger.json,swagger.yaml}` from `// @` annotations.
- **Run locally**: `docker compose up` — one container runs nginx + both hosts plus Vault, Postgres, IPFS, pgAdmin; console at `http://localhost:5500`, API at `http://localhost:5500/api`. Web dev mode: `cd web && npm run dev` (proxies `/api`).

Test-writing conventions live in the `go-*-test-conventions` skills — follow them (`when_<behavior>_test.go`, `Given.../Should...` nested `t.Run`, suite `BeforeEach`, gomock, testify, `model_builders`).
