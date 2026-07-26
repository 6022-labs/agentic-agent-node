-- +goose Up
-- add column "allowed_tool_patterns" to table: "mcp_connections"
-- sqlite requires a default when adding a NOT NULL column; "[]" allows every
-- tool, which is what pre-existing rows expect.
ALTER TABLE `mcp_connections` ADD COLUMN `allowed_tool_patterns` text NOT NULL DEFAULT '[]';

-- +goose Down
-- reverse: add column "allowed_tool_patterns" to table: "mcp_connections"
ALTER TABLE `mcp_connections` DROP COLUMN `allowed_tool_patterns`;
