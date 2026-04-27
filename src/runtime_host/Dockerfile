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

FROM scratch AS final

# Import the compiled executable from the first stage.
COPY --from=builder /app /app
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/migrations/postgres /migrations/postgres
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/migrations/mysql    /migrations/mysql
COPY --from=builder /usr/local/go/src/github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/migrations/mariadb  /migrations/mariadb
ENV AGENT_KNOWLEDGE_MIGRATIONS__FOLDER_PATH=/migrations

EXPOSE 5001

# Run the compiled binary at container start.
ENTRYPOINT ["/app"]
