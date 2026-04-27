FROM golang:1.26-alpine AS builder

WORKDIR /usr/local/go/src/github.com/6022-labs/agentic-agent-node

COPY ./go.mod ./
COPY ./go.sum ./

RUN go mod download

COPY ./ ./

# Build both binaries with the same flags (statically linked).
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /gateway ./src/gateway_host \
 && CGO_ENABLED=0 go build -installsuffix 'static' -o /runtime ./src/runtime_host

FROM alpine:3.20 AS final

# tini handles PID 1 + signal forwarding so SIGTERM reaches both child
# binaries cleanly. ca-certificates lets outbound TLS calls work (Vault, RPC,
# proxy.golang.org during dev).
RUN apk add --no-cache tini ca-certificates

COPY --from=builder /gateway /usr/local/bin/gateway
COPY --from=builder /runtime /usr/local/bin/runtime
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/migrations/postgres /migrations/postgres
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/migrations/mysql    /migrations/mysql
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/migrations/mariadb  /migrations/mariadb

# Pre-filled defaults that hold whenever the two binaries share a container:
# both bind the loopback interface and the gateway can reach the runtime at it.
# Anything sensitive (Vault token, JWT signing key, chain RPC, agent DB URL)
# must still be supplied at runtime.
ENV GATEWAY_HOST__API_PORT=5000 \
    GATEWAY_HOST__APP_NAME=AgenticAI-Agent-Node-Gateway \
    GATEWAY_HOST__BODY_LIMIT_BYTES=4194304 \
    RUNTIME_HOST__API_PORT=5001 \
    RUNTIME_HOST__APP_NAME=AgenticAI-Agent-Node-Runtime \
    RUNTIME_HOST__BODY_LIMIT_BYTES=4194304 \
    RUNTIME_HOST__LISTEN_ADDRESS=127.0.0.1 \
    RUNTIME_CLIENT__BASE_URL=http://127.0.0.1:5001 \
    RUNTIME_CLIENT__TIMEOUT=5s \
    JWT__TOKEN_TTL=30m \
    JWT__NONCE_TTL=5m \
    VAULT__KV_V1_MOUNT_PATH=secret \
    AGENT_KNOWLEDGE_MIGRATIONS__FOLDER_PATH=/migrations

# Entrypoint script: start runtime in background, then gateway in foreground.
# Tini reaps and forwards signals; if either child dies the script exits and
# the container restarts (per orchestrator restart policy).
COPY <<'EOF' /usr/local/bin/run.sh
#!/bin/sh
set -eu

cleanup() {
    [ -n "${RUNTIME_PID:-}" ] && kill -TERM "$RUNTIME_PID" 2>/dev/null || true
    wait "${RUNTIME_PID:-}" 2>/dev/null || true
}
trap cleanup TERM INT

runtime &
RUNTIME_PID=$!

# Give the runtime a moment to bind 127.0.0.1:$RUNTIME_HOST__API_PORT before
# the gateway starts dialling /internal/* on it.
sleep 1

gateway &
GATEWAY_PID=$!

# Exit as soon as either child exits; tini will SIGTERM the survivor.
wait -n "$RUNTIME_PID" "$GATEWAY_PID"
EXIT_CODE=$?
cleanup
exit $EXIT_CODE
EOF
RUN chmod +x /usr/local/bin/run.sh

EXPOSE 5000

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/run.sh"]
