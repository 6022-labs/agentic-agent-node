-- +goose Up
-- modify "mcp_connections" table
-- longtext cannot carry a DEFAULT, so existing rows are backfilled ("[]" allows
-- every tool) before the column becomes NOT NULL.
ALTER TABLE `mcp_connections` ADD COLUMN `allowed_tool_patterns` longtext NULL AFTER `url`;
UPDATE `mcp_connections` SET `allowed_tool_patterns` = '[]' WHERE `allowed_tool_patterns` IS NULL;
ALTER TABLE `mcp_connections` MODIFY COLUMN `allowed_tool_patterns` longtext NOT NULL;

-- +goose Down
-- reverse: modify "mcp_connections" table
ALTER TABLE `mcp_connections` DROP COLUMN `allowed_tool_patterns`;
