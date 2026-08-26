-- +goose Up
-- modify "conversations" table
ALTER TABLE `conversations` ADD COLUMN `name` varchar(128) NULL AFTER `agent_id`, ADD UNIQUE INDEX `uq_conversations_agent_name` (`agent_id`, `name`);

-- +goose Down
-- reverse: modify "conversations" table
ALTER TABLE `conversations` DROP INDEX `uq_conversations_agent_name`, DROP COLUMN `name`;
