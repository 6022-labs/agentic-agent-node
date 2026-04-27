-- +goose Up
-- create "conversation_messages" table
CREATE TABLE "public"."conversation_messages" (
  "id" text NOT NULL,
  "conversation_id" text NULL,
  "author" text NULL,
  "role" text NULL,
  "platform" text NULL,
  "content" text NULL,
  "hash" text NULL,
  "timestamp" timestamptz NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create "llm_connections" table
CREATE TABLE "public"."llm_connections" (
  "id" text NOT NULL,
  "platform" text NULL,
  "model" text NULL,
  "base_url" text NULL,
  "setup_status" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create "mcp_connections" table
CREATE TABLE "public"."mcp_connections" (
  "id" text NOT NULL,
  "name" text NULL,
  "url" text NULL,
  "header_auth_keys" jsonb NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);

-- +goose Down
-- reverse: create "mcp_connections" table
DROP TABLE "public"."mcp_connections";
-- reverse: create "llm_connections" table
DROP TABLE "public"."llm_connections";
-- reverse: create "conversation_messages" table
DROP TABLE "public"."conversation_messages";
