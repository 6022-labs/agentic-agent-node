-- +goose Up
-- create "conversation_messages" table — local agent transcript per conversation,
-- deduped by SHA-512 hash of (role, author, content, platform, conversationId).
CREATE TABLE "public"."conversation_messages" (
  "id"              uuid         NOT NULL,
  "conversation_id" varchar(128) NOT NULL,
  "author"          varchar(128) NOT NULL,
  "role"            varchar(32)  NOT NULL,
  "platform"        varchar(64),
  "content"         text         NOT NULL,
  "hash"            varchar(128) NOT NULL,
  "timestamp"       timestamptz  NOT NULL,
  "created_at"      timestamptz  NOT NULL DEFAULT now(),
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "uq_conversation_messages_hash"
  ON "public"."conversation_messages" ("hash");
CREATE INDEX "idx_conversation_messages_conversation_id"
  ON "public"."conversation_messages" ("conversation_id");

-- +goose Down
DROP INDEX IF EXISTS "public"."idx_conversation_messages_conversation_id";
DROP INDEX IF EXISTS "public"."uq_conversation_messages_hash";
DROP TABLE IF EXISTS "public"."conversation_messages";
