FROM node:26-alpine AS web-builder

WORKDIR /web

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./

# Vite inlines these at build time. The API base URL stays relative: nginx in
# this same image routes /api to the gateway, whatever host serves the bundle.
ARG VITE_GATEWAY_API_BASE_URL=/api
ARG VITE_APP_KIT_PROJECT_ID=6e030f5fb755b144ec2cb0d10f2572ae
ARG VITE_SCANNER_BASE_URL=https://amoy.polygonscan.com
ARG VITE_ENVIRONMENT=production
ARG VITE_DOCS_BASE_URL=https://docs.agentic.6022.io
ENV VITE_GATEWAY_API_BASE_URL=$VITE_GATEWAY_API_BASE_URL \
    VITE_APP_KIT_PROJECT_ID=$VITE_APP_KIT_PROJECT_ID \
    VITE_SCANNER_BASE_URL=$VITE_SCANNER_BASE_URL \
    VITE_ENVIRONMENT=$VITE_ENVIRONMENT \
    VITE_DOCS_BASE_URL=$VITE_DOCS_BASE_URL

RUN npm run build

FROM golang:1.27-alpine AS builder

WORKDIR /usr/local/go/src/github.com/6022-labs/agentic-agent-node

COPY ./go.mod ./
COPY ./go.sum ./

RUN go mod download

COPY ./ ./

# Build both binaries with the same flags (statically linked).
RUN CGO_ENABLED=0 go build -installsuffix 'static' -o /gateway ./src/gateway_host \
 && CGO_ENABLED=0 go build -installsuffix 'static' -o /runtime ./src/runtime_host

FROM alpine:3.24 AS final

# tini handles PID 1 + signal forwarding so SIGTERM reaches the children
# cleanly. ca-certificates lets outbound TLS calls work (Vault, RPC,
# proxy.golang.org during dev). nginx is the single public listener: web
# console at /, gateway proxied behind /api, /.well-known reserved.
# bash backs the run_command self tool — models write bashisms busybox sh chokes on.
RUN apk add --no-cache tini ca-certificates nginx bash

# Commands drop to this unprivileged account, so the agent cannot signal the
# runtime, the gateway or nginx — nor read root-owned /proc/*/environ.
RUN adduser -D -H -u 10001 agent \
 && mkdir -p /workspace \
 && chown agent:agent /workspace
ENV OS__COMMAND_USER=agent

COPY --from=builder /gateway /usr/local/bin/gateway
COPY --from=builder /runtime /usr/local/bin/runtime
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/runtime_db/migrations/postgres /migrations/postgres
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/runtime_db/migrations/mysql    /migrations/mysql
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/runtime_db/migrations/mariadb  /migrations/mariadb
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/runtime_db/migrations/sqlite   /migrations/sqlite
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/runtime_host/images/nft_mark.png /images/nft_mark.png
COPY --from=web-builder /web/dist /web
COPY nginx.conf /etc/nginx/http.d/default.conf

# Client IP for the rate limiter: nginx resolves it from the edge in front, trusting loopback, private and shared ranges.
ENV PROXY__HEADER=X-Forwarded-For \
    PROXY__TRUSTED_PROXIES=127.0.0.1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,100.64.0.0/10,fc00::/7

# Pre-filled defaults that hold whenever the three processes share a container:
# both Go hosts bind the loopback interface (only nginx is reachable from
# outside) and the gateway can reach the runtime at it. Anything sensitive
# (Vault token, JWT signing key, chain RPC, agent DB URL) must still be
# supplied at runtime.
ENV GATEWAY_HOST__API_PORT=5001 \
    GATEWAY_HOST__BODY_LIMIT_BYTES=4194304 \
    GATEWAY_HOST__LISTEN_ADDRESS=127.0.0.1 \
    GATEWAY__PUBLIC_PREFIX=/api \
    RUNTIME_HOST__API_PORT=5400 \
    RUNTIME_HOST__BODY_LIMIT_BYTES=4194304 \
    RUNTIME_HOST__LISTEN_ADDRESS=127.0.0.1 \
    RUNTIME_CLIENT__BASE_URL=http://127.0.0.1:5400 \
    RUNTIME_CLIENT__TIMEOUT=5 \
    JWT__TOKEN_TTL=1800 \
    JWT__NONCE_TTL=300 \
    VAULT__KV_V1_MOUNT_PATH=secret \
    IMAGE_GENERATION__IMAGE_MARK_PATH=/images/nft_mark.png \
    RUNTIME_MIGRATIONS__FOLDER_PATH=/migrations

# Entrypoint script: runtime first, then gateway, then nginx in front.
# Tini reaps and forwards signals; if any child dies the script exits and
# the container restarts (per orchestrator restart policy).
COPY <<'EOF' /usr/local/bin/run.sh
#!/bin/sh
set -eu

cleanup() {
    for PID in "${NGINX_PID:-}" "${GATEWAY_PID:-}" "${RUNTIME_PID:-}"; do
        [ -n "$PID" ] && kill -TERM "$PID" 2>/dev/null || true
    done
    for PID in "${NGINX_PID:-}" "${GATEWAY_PID:-}" "${RUNTIME_PID:-}"; do
        [ -n "$PID" ] && wait "$PID" 2>/dev/null || true
    done
}
trap 'cleanup; exit 143' TERM INT

runtime &
RUNTIME_PID=$!

# Give the runtime a moment to bind 127.0.0.1:$RUNTIME_HOST__API_PORT before
# the gateway starts dialling /internal/* on it.
sleep 1

# The gateway only sees nginx, so it reads the client IP nginx resolved from PROXY__*.
PROXY__HEADER=X-Real-IP PROXY__TRUSTED_PROXIES=127.0.0.1 gateway &
GATEWAY_PID=$!

REAL_IP_CONF=/etc/nginx/http.d/real_ip.conf
rm -f "$REAL_IP_CONF"
if [ -n "${PROXY__HEADER:-}" ]; then
    for PROXY in $(echo "${PROXY__TRUSTED_PROXIES:-}" | tr ',' ' '); do
        echo "set_real_ip_from $PROXY;" >> "$REAL_IP_CONF"
    done
    echo "real_ip_header $PROXY__HEADER;" >> "$REAL_IP_CONF"
    echo "real_ip_recursive on;" >> "$REAL_IP_CONF"
fi

nginx -g 'daemon off;' &
NGINX_PID=$!

# Exit as soon as any child dies. Poll instead of `wait -n pid...`:
# busybox ash ignores the PID list, so an already-dead child never wakes it.
while kill -0 "$RUNTIME_PID" 2>/dev/null \
   && kill -0 "$GATEWAY_PID" 2>/dev/null \
   && kill -0 "$NGINX_PID" 2>/dev/null; do
    sleep 2
done
cleanup
exit 1
EOF
RUN chmod +x /usr/local/bin/run.sh

EXPOSE 5000

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/run.sh"]
