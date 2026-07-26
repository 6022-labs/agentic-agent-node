-- +goose Up
-- modify "mcp_connections" table
-- The default backfills existing rows ("[]" allows every tool) and is dropped
-- afterwards so the column matches the gorm model.
ALTER TABLE "public"."mcp_connections" ADD COLUMN "allowed_tool_patterns" text NOT NULL DEFAULT '[]';
ALTER TABLE "public"."mcp_connections" ALTER COLUMN "allowed_tool_patterns" DROP DEFAULT;

-- +goose Down
-- reverse: modify "mcp_connections" table
ALTER TABLE "public"."mcp_connections" DROP COLUMN "allowed_tool_patterns";
