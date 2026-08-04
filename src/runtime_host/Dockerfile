FROM golang:1.26-alpine AS builder

WORKDIR /usr/local/go/src/github.com/6022-labs/agentic-agent-node

COPY ./go.mod ./
COPY ./go.sum ./

RUN go mod download

COPY ./ ./

WORKDIR /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/runtime_host

# Build the executable to `/app`. Mark the build as statically linked.
RUN CGO_ENABLED=0 go build \
    -installsuffix 'static' \
    -o /app .

FROM alpine:3.20 AS final

# bash backs the run_command self tool — models write bashisms busybox sh chokes on. tini reaps the orphans a killed
# process group leaves behind, which the Go binary would not do as PID 1.
RUN apk add --no-cache ca-certificates bash tini

# Import the compiled executable from the first stage.
COPY --from=builder /app /app
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/runtime_db/migrations/postgres /migrations/postgres
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/runtime_db/migrations/mysql    /migrations/mysql
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/runtime_db/migrations/mariadb  /migrations/mariadb
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/runtime_db/migrations/sqlite   /migrations/sqlite
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/runtime_host/images/nft_mark.png /images/nft_mark.png
ENV IMAGE_GENERATION__IMAGE_MARK_PATH=/images/nft_mark.png \
    RUNTIME_MIGRATIONS__FOLDER_PATH=/migrations

# Commands drop to this unprivileged account, so the agent cannot signal the
# runtime, the gateway or nginx — nor read root-owned /proc/*/environ.
RUN adduser -D -H -u 10001 agent \
 && mkdir -p /workspace \
 && chown agent:agent /workspace
ENV OS__COMMAND_USER=agent

EXPOSE 5400

# Run the compiled binary at container start.
ENTRYPOINT ["/sbin/tini", "--", "/app"]
