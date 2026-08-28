-- +goose Up
-- add column "name" to table: "conversations"
ALTER TABLE `conversations` ADD COLUMN `name` text NULL;
-- create index "uq_conversations_agent_name" to table: "conversations"
CREATE UNIQUE INDEX `uq_conversations_agent_name` ON `conversations` (`agent_id`, `name`);

-- +goose Down
-- reverse: create index "uq_conversations_agent_name" to table: "conversations"
DROP INDEX `uq_conversations_agent_name`;
-- reverse: add column "name" to table: "conversations"
ALTER TABLE `conversations` DROP COLUMN `name`;
