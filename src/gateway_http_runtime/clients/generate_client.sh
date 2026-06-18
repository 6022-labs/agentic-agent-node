#!/usr/bin/env sh
set -eu

# Regenerates the runtime HTTP client from runtime_host's OpenAPI v3 spec.
# Run src/runtime_host/docs/generate_docs.sh first if the runtime API changed.

CLIENTS_DIR="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
SPEC_PATH="${CLIENTS_DIR}/../../runtime_host/docs/swagger.v3.yaml"
OUTPUT_FILE="${CLIENTS_DIR}/oapi_client.gen.go"
SANITIZED_SPEC="$(mktemp -t runtime-spec-XXXXXX).yaml"
trap 'rm -f "${SANITIZED_SPEC}"' EXIT

OAPI_CODEGEN_VERSION="v2.5.1"

# swag v2 emits the Swagger-2.0 `type: file` for multipart uploads, which
# oapi-codegen rejects on a 3.1 spec. Rewrite it to the 3.x binary-string form.
echo "[1/2] Sanitizing multipart file types for oapi-codegen"
perl -0pe 's/(\s*)type: file/${1}type: string${1}format: binary/g' \
	<"${SPEC_PATH}" >"${SANITIZED_SPEC}"

echo "[2/2] Generating runtime client with oapi-codegen ${OAPI_CODEGEN_VERSION}"
(
	cd "${CLIENTS_DIR}"
	go run "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@${OAPI_CODEGEN_VERSION}" \
		-generate types,client \
		-package clients \
		-o "${OUTPUT_FILE}" \
		"${SANITIZED_SPEC}"
)

echo "Done"
