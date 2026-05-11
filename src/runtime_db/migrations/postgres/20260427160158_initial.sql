-- +goose Up
-- create "conversation_messages" table
CREATE TABLE "public"."conversation_messages" (
  "id" character varying(36) NOT NULL,
  "conversation_id" character varying(128) NOT NULL,
  "role" character varying(32) NOT NULL,
  "content" text NOT NULL,
  "hash" character varying(128) NOT NULL,
  "timestamp" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_conversation_messages_conversation_id" to table: "conversation_messages"
CREATE INDEX "idx_conversation_messages_conversation_id" ON "public"."conversation_messages" ("conversation_id");
-- create index "uq_conversation_messages_hash" to table: "conversation_messages"
CREATE UNIQUE INDEX "uq_conversation_messages_hash" ON "public"."conversation_messages" ("hash");
-- create "llm_connections" table
CREATE TABLE "public"."llm_connections" (
  "id" character varying(36) NOT NULL,
  "platform" character varying(64) NOT NULL,
  "model" character varying(255) NOT NULL,
  "base_url" text NULL,
  "setup_status" character varying(32) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- create "mcp_connections" table
CREATE TABLE "public"."mcp_connections" (
  "id" character varying(36) NOT NULL,
  "name" character varying(128) NOT NULL,
  "url" text NOT NULL,
  "header_auth_keys" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);

-- +goose Down
-- reverse: create "mcp_connections" table
DROP TABLE "public"."mcp_connections";
-- reverse: create "llm_connections" table
DROP TABLE "public"."llm_connections";
-- reverse: create index "uq_conversation_messages_hash" to table: "conversation_messages"
DROP INDEX "public"."uq_conversation_messages_hash";
-- reverse: create index "idx_conversation_messages_conversation_id" to table: "conversation_messages"
DROP INDEX "public"."idx_conversation_messages_conversation_id";
-- reverse: create "conversation_messages" table
DROP TABLE "public"."conversation_messages";
