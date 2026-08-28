-- +goose Up
-- modify "conversations" table
ALTER TABLE "public"."conversations" ADD COLUMN "name" character varying(128) NULL;
-- create index "uq_conversations_agent_name" to table: "conversations"
CREATE UNIQUE INDEX "uq_conversations_agent_name" ON "public"."conversations" ("agent_id", "name");

-- +goose Down
-- reverse: create index "uq_conversations_agent_name" to table: "conversations"
DROP INDEX "public"."uq_conversations_agent_name";
-- reverse: modify "conversations" table
ALTER TABLE "public"."conversations" DROP COLUMN "name";
