-- +goose Up
-- add column "caller_id" to table: "conversations"
ALTER TABLE `conversations` ADD COLUMN `caller_id` text NULL;
-- backfill: pre-existing sessions become owner-only (the agent's owner of record); rows without one stay console-only
UPDATE `conversations` SET `caller_id` = (SELECT LOWER(`agents`.`owner`) FROM `agents` WHERE `agents`.`id` = `conversations`.`agent_id`) WHERE `caller_id` IS NULL;
-- drop index "uq_conversations_agent_name" from table: "conversations"
DROP INDEX `uq_conversations_agent_name`;
-- create index "uq_conversations_agent_caller_name" to table: "conversations"
CREATE UNIQUE INDEX `uq_conversations_agent_caller_name` ON `conversations` (`agent_id`, `caller_id`, `name`);

-- +goose Down
-- reverse: create index "uq_conversations_agent_caller_name" to table: "conversations"
DROP INDEX `uq_conversations_agent_caller_name`;
-- reverse: drop index "uq_conversations_agent_name" from table: "conversations"
CREATE UNIQUE INDEX `uq_conversations_agent_name` ON `conversations` (`agent_id`, `name`);
-- reverse: add column "caller_id" to table: "conversations"
ALTER TABLE `conversations` DROP COLUMN `caller_id`;
