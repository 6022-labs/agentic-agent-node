# Gateway host — API docs

Swagger 2.0 + OpenAPI 3.1 specs are generated from `@Summary`/`@Tags`/`@Router`/etc. annotations on the gateway_mvc controllers.

## Regenerate

```sh
./generate_docs.sh
```

Writes:
- `docs.go` — Go-embedded spec used by the `_ "github.com/.../docs"` blank import in `server_configuration.go`.
- `swagger.json` / `swagger.yaml` — Swagger 2.0.
- `swagger.v3.json` / `swagger.v3.yaml` — OpenAPI 3.1.

## View

When the gateway is running directly (`go run ./src/gateway_host`):

```
http://localhost:5000/swagger/index.html
```

Through the container's nginx (docker compose) the UI is at
`http://localhost:5500/api/swagger/index.html` (nginx strips `/api`). Note
"Try it out" targets the origin root there, so use the direct gateway URL
for interactive calls.
