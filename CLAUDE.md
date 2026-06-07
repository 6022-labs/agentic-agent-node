# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Code style

- **Keep comments short.** One line where one line does. State the non-obvious *why* or a cross-package contract; never restate what the code already says. No multi-paragraph doc blocks. Go doc comments start with the identifier name, then stop.
- **Document the contract once, on the interface.** Method doc comments live on the port (`*Interface`); the implementation's methods carry **no** doc comment (it would only drift). Keep the type-level comment on the impl and comments on private helpers. See `RegistrationReconciler` / `AgentCollectionRequester`.
- **Name dependency fields/params after their full type** (the mvc-controller convention), not a generic role. Drop the package qualifier and the `Interface` suffix, camelCase the rest: `EnsPublisherInterface` → `ensPublisher` (not `publisher`), `AgentRegistrationStateRepositoryInterface` → `agentRegistrationStateRepository` (not `stateRepository`), `*settings.AgentIdentitySettings` → `agentIdentitySettings`. Idiomatic infra names (`logger`, `client`, `ctx`) and local *domain-value* variables (`agent`, `state`, `collection`, `tokenId`) keep their short names.
- Hexagonal layering is enforced by structure, not lint — keep domain logic out of adapters (see Architecture).

## Commands

```bash
go build ./src/...                                   # build everything
find . -type d -name '*unit_tests' -exec go test {}/... \;   # all unit tests (CI does this)
go test ./tests/<package>_unit_tests/...             # one package
```

- **Integration tests** (`tests/*_integration_tests`) spin up real Postgres/Redis via `tests/common_tests/container_builders` — they need Docker, and are NOT in the unit-test CI sweep.
- **Mocks** are generated, not hand-written. Regenerate after changing any interface: `deepsearch-mockgen -S ./src -O ./tests -A -P` (CLI: https://github.com/FournyP/deepsearch-mockgen-cli).
- **Migrations**: `src/runtime_db/migrations/generate_migrations.sh` (atlas+gorm, per-driver SQL under `postgres/`, `mysql/`, `mariadb/`).
- **Swagger docs**: `./scripts/generate_swagger_docs_for_hosts.sh [--host gateway_host]` — regenerates `*_host/docs/{docs.go,swagger.json,swagger.yaml}` from `// @` annotations in host `main.go` and controllers.
- **Run locally**: `docker compose up` — one container runs both hosts plus Vault, Postgres, pgAdmin.

Conventions for writing tests live in `.claude/skills/go-*-test-conventions` — follow them (`when_<behavior>_test.go`, `Given.../Should...` nested `t.Run`, suite `BeforeEach`, gomock, testify, `model_builders`).

## Architecture

Single Go module, **two binaries** built from one image:

- **`gateway_host`** (public) — the only externally-reachable surface. Wallet auth (challenge nonce → EIP-191 `personal_sign` → JWT), config endpoints, `/responses`. Authenticated routes proxy over the loopback to runtime_host via `gateway_http_runtime`.
- **`runtime_host`** (internal, bound to 127.0.0.1) — everything under `/internal`: inference loop, knowledge persistence, MCP/LLM config, identity. Owns wallet bootstrap (runs synchronously at startup before serving).

One agent = one container. The runtime DB therefore holds **zero or one** LLM connection row.

**Module layout** — each domain module is a directory under `src/` with `configurations/ services/ services/interfaces/ repositories/ models/ settings/ use_cases/`. Ports are the `interfaces/` packages; adapters are separate modules that implement them:

| Domain | HTTP/infra adapters |
|--------|--------|
| `agent_reasoning` (inference loop, MCP + LLM connections) | `agent_reasoning_http_inference` (per-provider LLM clients: anthropic/openai/bedrock/ollama/...), `agent_reasoning_http_mcp` (MCP tool calls) |
| `agent_identity` (on-chain NFT identity) | `agent_identity_blockchain` (ERC721 collections via AgentCollectionsManager), `agent_identity_ens` (AgentEnsRegistry) |
| `agent_knowledge` (conversation transcripts) | `runtime_db` (gorm + goose migrations) |
| `gateway` (auth, config proxy) | `gateway_http_runtime` (HTTP client to runtime), `gateway_mvc` (fiber controllers) |

**Cross-cutting:**
- **DI**: `go.uber.org/dig`. Each module exposes `Add<Module>Configuration(container)`, all wired in the host's `configurations/di_configuration.go`. Add a new module by adding its `Add...` call there.
- **Config**: `koanf` — JSON `appsettings` + `.env`/env-var overrides (double-underscore nesting, e.g. `JWT__SIGNING_KEY`). Settings structs live in each module's `settings/`.
- **Secrets** never touch the DB or config rows — `runtime_vault` (HashiCorp Vault) holds LLM api keys, wallet credentials, and MCP header-auth values. Rows store only the *key/header names*.
- **Use-case results**: use cases return idiomatic `(*response, error)` (or just `error` for no-body operations). Failures are concrete error types from `common/errors` (`ValidationError`→400, `UnauthorizedError`→401, `NotFoundError`→404, `ConflictError`→409, `UnavailableError`→502, `InternalError`→500); a plain error is treated as 500. Controllers map the error type to an HTTP status (each `*_mvc` package has a `writeError` helper) and choose the success status (200/201/204). Don't return fiber/HTTP types from use cases.
- **Identity resolves lazily** — `AgentIdentity` is absent until an NFT is minted; `ErrIdentityNotFound` is the shared sentinel both the in-process resolver and the gateway HTTP client return.
