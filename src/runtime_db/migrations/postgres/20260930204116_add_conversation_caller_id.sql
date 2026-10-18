-- +goose Up
-- modify "conversations" table
ALTER TABLE "public"."conversations" ADD COLUMN "caller_id" character varying(128) NULL;
-- backfill: pre-existing sessions become owner-only (the agent's owner of record); rows without one stay console-only
UPDATE "public"."conversations" SET "caller_id" = LOWER("agents"."owner") FROM "public"."agents" WHERE "agents"."id" = "conversations"."agent_id" AND "conversations"."caller_id" IS NULL AND "agents"."owner" IS NOT NULL;
-- drop index "uq_conversations_agent_name" from table: "conversations"
DROP INDEX "public"."uq_conversations_agent_name";
-- create index "uq_conversations_agent_caller_name" to table: "conversations"
CREATE UNIQUE INDEX "uq_conversations_agent_caller_name" ON "public"."conversations" ("agent_id", "caller_id", "name");

-- +goose Down
-- reverse: create index "uq_conversations_agent_caller_name" to table: "conversations"
DROP INDEX "public"."uq_conversations_agent_caller_name";
-- reverse: drop index "uq_conversations_agent_name" from table: "conversations"
CREATE UNIQUE INDEX "uq_conversations_agent_name" ON "public"."conversations" ("agent_id", "name");
-- reverse: modify "conversations" table
ALTER TABLE "public"."conversations" DROP COLUMN "caller_id";
