# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Code style

### Comments — be sharp, write fewer

You write too many comments. Cut hard. The defaults:

- **One line, max.** A comment is one line. The _only_ exception is an interface or interface method, which may use **two**. Never write a multi-line or multi-paragraph doc block on anything.
- **Only say the non-obvious.** State a _why_, an invariant, or a cross-package contract. Never restate what the code already says. If the code is clear, write **no** comment — that is the common case.
- **Go doc form.** Doc comments start with the identifier name, then stop (`AgentMinter mints ...`).
- **Document the contract once, on the port.** Method doc comments live on the `*Interface`; the implementation's methods carry **no** doc comment (it only drifts). The impl keeps its type-level comment and comments on private helpers. See `RegistrationReconciler` / `AgentCollectionRequester`.

### Naming

- **Name dependency fields/params after their full type** (the mvc-controller convention), not a generic role. Drop the package qualifier and `Interface` suffix, camelCase the rest: `EnsPublisherInterface` → `ensPublisher` (not `publisher`), `AgentRegistrationStateRepositoryInterface` → `agentRegistrationStateRepository` (not `stateRepository`), `*settings.AgentIdentitySettings` → `agentIdentitySettings`. Idiomatic infra names (`logger`, `client`, `ctx`) and local _domain-value_ variables (`agent`, `state`, `collection`, `tokenId`) keep their short names.

### Helpers hang off a type, not the package

A private helper belongs on the type that uses it, as a method — not as a package-level `func`. A free function is visible to every file in its package, so it invites name collisions and reads as shared API when it is really one type's internals.

- `func (r *SignX402PaymentRequest) extraString(key string) string` — yes.
- `func extraString(extra map[string]any, key string) string` — no.

The exceptions, and they are the only ones: `New...` constructors; `Add<Module>Configuration` and its `provide...` helpers in a `configurations/` package; enum helpers like `GetAll...`; free functions in a `utils/` or `helpers/` package; and the exported API of a deliberate utility package (`common/x402`). Package-level `var` is fine for an immutable value a `const` cannot express (a sentinel error, a `*big.Int`, a lookup table).

## Architecture

Single Go module (`github.com/6022-labs/agentic-agent-node`, Go 1.26), **two binaries** + the web console, all in one image. **One agent = one container** — so the runtime DB holds at most **one** LLM connection row and **one** agent-identity row.

### Hexagonal (ports & adapters) — enforced by structure, not lint

Each **domain module** is a directory under `src/` with `configurations/ services/ services/interfaces/ repositories/ models/ settings/ use_cases/`. The **ports** are the `interfaces/` packages (service ports under `services/interfaces/`, persistence ports under `repositories/`). The **adapters** are _separate sibling modules_ that implement those ports (`*_blockchain`, `*_http_*`, `*_mvc`, `runtime_db`, `runtime_vault`).

Rules that the layout exists to enforce — keep them:

- **Domain logic lives in the domain module only.** Adapters translate (HTTP/RPC/SQL/Vault ↔ domain types) and do nothing else. No business rules, no orchestration, no validation-with-meaning in an adapter.
- **Inbound adapters are transport only — controllers AND listeners.** A `*_mvc` controller, a blockchain/event listener, or a queue consumer all just _enter_ the business logic: decode/deserialize → call exactly one use case → map the result back (status via `common/mvc.WriteError`, ack/nack, etc.). **No business logic and no validation rules in them** — message/body/field validation carries meaning and belongs in the use case, not the transport. If a decision is being made in a controller or listener, it belongs in a use case. Don't return fiber/HTTP types from use cases.
- **Use cases are the entry points to the domain.** They return idiomatic `(*response, error)` (or just `error` for no-body ops) and depend only on ports.
- Adding a behavior usually means: define/extend a port in `services/interfaces/`, implement the domain logic in `services/` or `use_cases/`, implement the port in the relevant adapter module, wire it in DI.

### Clear boundaries — isolate tier-services, not everything

A **tier-service** is any dependency that crosses a boundary: it reaches an external system, does I/O, or is a side-effecting / cross-cutting infra concern. Each has **exactly one owner package + an interface**; nothing else imports the underlying library — callers depend on the port and mock it. **Pure utility libraries are not boundaries** (`samber/lo`, stdlib, math) — import them anywhere; don't wrap them. Litmus: _external system / I/O / side effect / cross-cutting tier_ → isolate behind a service + port; _pure in-memory computation_ → use freely.

| Tier-service                     | Sole owner                                                                                                   | Everyone else uses                                                |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------- |
| Metrics / traces (OpenTelemetry) | each module's `metrics/` + `traces/` sub-package (+ `common/host_configurations/telemetry_configuration.go`) | the local `metrics`/`traces` `interfaces/` port                   |
| Persistence (gorm)               | `*_db` (here `runtime_db`)                                                                                   | repository ports; models carry tags only                          |
| HTTP transport (fiber)           | `*_mvc`, `common/mvc`, `*_host`                                                                              | `(*response, error)` from use cases — never fiber types           |
| Blockchain RPC (go-ethereum)     | `*_blockchain` (here `agent_identity_blockchain`, `agent_payment_blockchain`; + `runtime_common` RPC infra)  | service/repository ports (address/hash value-types excepted)      |
| Secrets (Vault)                  | `*_vault` (here `runtime_vault`)                                                                             | credential-store ports (e.g. `WalletCredentialsStorageInterface`) |
| IPFS                             | `agent_identity_http_ipfs`                                                                                   | port                                                              |
| Local machine (`os`, `os/exec`, `syscall`) | `agent_reasoning_os`                                                                               | `CommandRunnerInterface` / `CommandHandleInterface`               |
| LLM / MCP SDKs                   | `agent_reasoning_http_inference`, `agent_reasoning_http_mcp`                                                 | reasoning ports                                                   |

A tier-service import found outside its owner is a boundary leak — fix it, don't spread it.

**Metrics & traces — the example to copy.** otel is touched in exactly one place per module: a dedicated `metrics/` (and, where needed, `traces/`) sub-package wrapping the meter/tracer and exposing a `…/interfaces` port (`MintMetrics` → `MintMetricsInterface`), injected via dig and mocked in `tests/<module>_mocks/metrics_mocks/`. **No service, use case, or controller imports `go.opentelemetry.io`.** New metric/trace = add a method to the module's `metrics/`/`traces/` package + its interface, then regenerate mocks. (Constructors fall back to a noop meter so tests never panic.)

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

| Domain (`src/<domain>`) | Purpose                                                                                                                               | Adapter modules implementing its ports                                                                                                                                                                                                                                                                                      |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `agent_reasoning`       | inference loop, LLM + MCP connections                                                                                                 | `agent_reasoning_http_inference` (per-provider LLM clients: anthropic/openai/bedrock/ollama/…), `agent_reasoning_http_mcp` (MCP tool calls), `agent_reasoning_os` (commands on the agent's own machine: `os/exec`, process-group kill, capped output, background-process registry), `agent_reasoning_mvc` (controllers), `runtime_db` (connection rows), `runtime_vault` (api keys, MCP header-auth) |
| `agent_identity`        | on-chain NFT identity, ENS, wallet bootstrap                                                                                          | `agent_identity_blockchain` (ERC721 mint via AgentCollectionsManager, ENS text records via AgentEnsRegistry, $6022 token reads; shared RPC/transactor/signing infra; all contract bindings in `abi`), `agent_identity_http_ipfs` (image/metadata to IPFS), `agent_identity_mvc`, `runtime_db`, `runtime_vault` (wallet key) |
| `agent_knowledge`       | conversation transcripts                                                                                                              | `agent_knowledge_mvc`, `runtime_db`                                                                                                                                                                                                                                                                                         |
| `agent_payment`         | payment policy + rules, x402 protocol                                                                                                 | `agent_payment_blockchain` (on-chain ERC-20 validation), `agent_payment_mvc`, `runtime_db`; settled via the x402 facilitator (see below)                                                                                                                                                                                    |
| `runtime_common`        | shared blockchain/operator config (chains, $6022 token, payment facilitators, collection managers, ENS registries) read from settings; cross-domain ports (`SecretStorageInterface`, `UnitOfWorkFactoryInterface`) | `runtime_common_mvc` (exposes blockchain config)                                                                                                                                                                                                                                                                            |
| `gateway`               | auth + config proxy                                                                                                                   | `gateway_http_runtime` (HTTP client to runtime), `gateway_http_x402` (x402 facilitator client), `gateway_mvc` (fiber controllers)                                                                                                                                                                                           |

### Payment (x402)

`agent_payment` is method-agnostic via a strategy pattern (`PaymentMethodStrategiesInterface` dispatches by `Method`); **x402** is the implemented method. An x402 rule (`PaymentReceiveRuleX402Config`) carries a CAIP-2 `network`, ERC-20 `asset` (default $6022), `payTo`, and `assetTransferMethod` (ERC-3009 `transferWithAuthorization` or permit2). **Zero-price rules never settle on-chain** — they carry only EIP-712 signing context. Actual settlement goes through an **x402 facilitator** (`gateway_http_x402`, long timeout because `/settle` blocks until the tx lands); facilitators are configured under `payment.receive.facilitators` (`runtime_common/settings/payment_receive_settings.go`) and a facilitator is only valid if a matching chain RPC exists in `blockchain.evm_chains`.

**Spending is gated per signature.** `X402PaymentSpendPolicyEvaluator.Authorize` runs inside `X402PaymentSigner.Sign`, ahead of the Permit2 approval so a refusal costs no gas. It denies by default: a quote needs an x402 `PaymentSpendRule` matching all four facets of `PaymentSpendRuleX402Config`, and the cap is the **minimum** `MaxAmount` over matching rules — a broad rule never widens a narrow one. `"0"` denies anything priced. `PaymentSpendPolicyProvider` seeds a fresh agent wide open (`.*`, max-uint256), so the gate binds only once an owner narrows it. Still unbounded: **spend rate**, the cap being per signature.

### Identity resolves lazily

`AgentIdentity` is absent until an NFT is minted; `ErrIdentityNotFound` is the shared sentinel returned by both the in-process resolver and the gateway HTTP client. `RegistrationReconciler` loads confirmed on-chain facts (owner, creator, token) and persists them into the agent + registration-state rows (used by check-mint and self-register).

## Cross-cutting

- **DI**: `go.uber.org/dig`. Each module exposes `Add<Module>Configuration(container)`; all are wired in the host's `configurations/di_configuration.go`. Add a module by adding its `Add...` call there. Controllers register into dig groups (`gateway_controllers` / `runtime_controllers`) and are injected as slices for route registration.
- **Config — every setting must be configurable from BOTH `appsettings` JSON AND env/`.env`.** `koanf` loads `appsettings*.json` then overlays env vars; nesting uses double-underscore lowercased to dotted keys (`JWT__SIGNING_KEY` → `jwt.signing_key`, `GATEWAY_HOST__API_PORT` → `gateway_host.api_port`). Settings structs live in each module's `settings/` with `koanf:"..."` tags and sane defaults in their `New...` constructor. When you add a setting, give it a tag, a default, and confirm the env override path works. Examples: `gateway_host/appsettings.example.json`, `runtime_host/appsettings.example.json`. (See the `koanf-settings` skill.)
- **Secrets never touch the DB or config rows** — `runtime_vault` (HashiCorp Vault, KV v1) holds LLM api keys, the agent wallet credentials, and MCP header-auth values. Rows store only the _key/header names_.
- **Errors → HTTP**: use cases fail with concrete types from `common/errors`, mapped by `common/mvc.WriteError`: `ValidationError`→400, `UnauthorizedError`→401, `NotFoundError`→404, `ConflictError`→409, `UnavailableError`→502, `InternalError`→500; a plain error → 500. Controllers pick the success status (200/201/204).
- **Persistence**: `runtime_db` is gorm; migrations are generated per driver (postgres/mysql/mariadb) via atlas+gorm.
- **Atomic writes**: several repository calls share one transaction through `UnitOfWorkFactoryInterface` (`runtime_common` port, `runtime_db/services` impl). Every call inside the scope takes `unitOfWork.Context()`; repositories join it via `GetDB(ctx)`. Vault and RPC calls never run inside a scope, and scopes stay short (sqlite has a single connection).

## Review rules — apply without being asked

### Source

- **`configurations/` is wiring only.** Any logic found there moves to `helpers/` or `services/` and gets its test there.
- **One owner per rule.** A rule shared by several requests lives in one base request that the others embed and call first; a single-field base is named after its field. Sibling requests validate the same field the same way.
- **Code lives in the narrowest module that uses it** (a host-only middleware stays in that host's module).
- **Docs match behavior.** Swagger declares every header a route reads and every status it can return, with the real error types; regenerate docs and clients after any change.
- **Adapters pass domain errors through unchanged**; only transport or unexpected failures are remapped.
- **Prefer a plain error to a clever fallback.**

### Tests

- **Test behavior where it lives, once.** No transport (controller, body binding) or `configurations/` tests; one file and one suite per SUT method; validation rules only in the request test, with the use case keeping a single invalid-request Given.
- **Structure**: nested `Given…` (conditions) / `Should…` (one outcome, no condition in the title). Given order follows the code: validation first, dependencies in call order, success last. Per-Given `initSuite` closure; request built at Given level.
- **Assert the outcome the title names** (error type and code, or the field), never a bare `assert.Error`.
- **One wording per situation, repo-wide.** Before writing a title, reuse the wording existing tests already use for the same situation; similar flows (a use case and the client it forwards to, sibling operations) use the same Givens in the same order.
- **No duplicate Givens.** Integration tests follow the `tests/runtime_db_integration_tests` style; no data-migration tests for now.
- **One `t.Run` per case**: no table-driven cases, loops or concatenated titles — except database integration tests, which loop over every supported database.
- **`BeforeEach` only builds the suite** (sut and mocks): it never captures values (no `Do`/`DoAndReturn` writing into suite fields); a case that checks a call asserts it with its own gomock expectation.
- **Valid Ethereum addresses are random**: build them with `common_tests.RandomEthAddress()`; only invalid addresses stay literal.

## Commands

```bash
go build ./src/...                                            # build everything (CI: go-build.yml)
find . -type d -name '*unit_tests' -exec go test {}/... \;   # all unit tests (CI: go-unit-tests.yml)
go test ./tests/<package>_unit_tests/...                     # one package
make fmt                                                      # golines, 120-char cap (needs golines)
```

**Always run `make fmt` before handing work back** — it wraps code lines to the 120-char cap and **errors on comments over the cap**: shorten those comments, never wrap them.

- **Integration tests** (`tests/*_integration_tests`) spin up real Postgres/Vault via `tests/common_tests/container_builders` — they need Docker and are **not** in the unit-test CI sweep.
- **Mocks** are generated, never hand-written. Regenerate after changing any interface: `deepsearch-mockgen -S ./src -O ./tests -A -P` (CLI: https://github.com/FournyP/deepsearch-mockgen-cli).
- **Migrations**: `src/runtime_db/migrations/generate_migrations.sh` (atlas+gorm; per-driver SQL under `postgres/`, `mysql/`, `mariadb/`).
- **Swagger docs**: `sh src/runtime_host/docs/generate_docs.sh` and `sh src/gateway_host/docs/generate_docs.sh` (or `make swagger` for both) — regenerates `*_host/docs/{docs.go,swagger.json,swagger.yaml}` from `// @` annotations. The runtime docs feed the generated gateway client, so regenerate them **before** `sh src/gateway_http_runtime/clients/generate_client.sh`.
- **Run locally**: `docker compose up` — one container runs nginx + both hosts plus Vault, Postgres, IPFS, pgAdmin; console at `http://localhost:5500`, API at `http://localhost:5500/api`. Web dev mode: `cd web && npm run dev` (proxies `/api`). Local chain: add `-f docker-compose.local-chain.yml` (Anvil `31337` + contracts deployer); its `EXPECTED_CONTRACTS` must match the `31337` entries in `src/runtime_common/settings/*.json`.

Test-writing conventions live in the `go-*-test-conventions` skills — follow them (`when_<behavior>_test.go`, `Given.../Should...` nested `t.Run`, suite `BeforeEach`, gomock, testify, `model_builders`).
