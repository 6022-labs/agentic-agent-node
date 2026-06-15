-- +goose Up
-- modify "agents" table
ALTER TABLE `agents` ADD COLUMN `chain_id` varchar(64) NOT NULL AFTER `id`;

-- +goose Down
-- reverse: modify "agents" table
ALTER TABLE `agents` DROP COLUMN `chain_id`;
