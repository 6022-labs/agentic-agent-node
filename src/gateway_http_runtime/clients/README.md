# gateway_http_runtime clients

`oapi_client.gen.go` is the generated HTTP transport the gateway uses to reach
`runtime_host`. It is produced by `oapi-codegen` (pinned `v2.5.1`,
`-generate types,client`) from `src/runtime_host/docs/swagger.v3.yaml`.

Do not edit `oapi_client.gen.go` by hand. To refresh after the runtime API changes:

```sh
sh src/runtime_host/docs/generate_docs.sh      # rebuild the runtime OpenAPI spec
sh src/gateway_http_runtime/clients/generate_client.sh
```

The typed `services/` in this module wrap the generated `ClientWithResponses`,
mapping the runtime wire shapes to/from the `External*` models in `../models`
(the gateway port contract). `../helpers/runtime_response.go` decodes responses
and translates non-2xx statuses into the shared `common/errors` types.
