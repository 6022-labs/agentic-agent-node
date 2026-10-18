-- +goose Up
-- modify "conversations" table
ALTER TABLE `conversations` ADD COLUMN `caller_id` varchar(128) NULL AFTER `agent_id`, DROP INDEX `uq_conversations_agent_name`, ADD UNIQUE INDEX `uq_conversations_agent_caller_name` (`agent_id`, `caller_id`, `name`);
-- backfill: pre-existing sessions become owner-only (the agent's owner of record); rows without one stay console-only
UPDATE `conversations` JOIN `agents` ON `agents`.`id` = `conversations`.`agent_id` SET `conversations`.`caller_id` = LOWER(`agents`.`owner`) WHERE `conversations`.`caller_id` IS NULL AND `agents`.`owner` IS NOT NULL;

-- +goose Down
-- reverse: modify "conversations" table
ALTER TABLE `conversations` DROP INDEX `uq_conversations_agent_caller_name`, ADD UNIQUE INDEX `uq_conversations_agent_name` (`agent_id`, `name`), DROP COLUMN `caller_id`;
