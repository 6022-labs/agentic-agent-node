# Runtime host — API docs

Swagger 2.0 + OpenAPI 3.1 specs for the internal `/internal/*` surface, generated from runtime_mvc annotations.

## Regenerate

```sh
./generate_docs.sh
```

Writes:
- `docs.go` — Go-embedded spec used by the `_ "github.com/.../docs"` blank import in `server_configuration.go`.
- `swagger.json` / `swagger.yaml` — Swagger 2.0.
- `swagger.v3.json` / `swagger.v3.yaml` — OpenAPI 3.1.

## View

Runtime binds 127.0.0.1 only, so this is reachable from within the container or via SSH tunnel:

```
http://127.0.0.1:5001/swagger/index.html
```
