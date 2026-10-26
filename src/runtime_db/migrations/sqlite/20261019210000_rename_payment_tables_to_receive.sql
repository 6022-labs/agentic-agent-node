-- +goose Up
-- Renames only, hand-written: atlas diffs a rename as DROP+CREATE, which would
-- destroy every deployed agent's pricing rules. SQLite rewrites child foreign-key
-- references on RENAME TO but cannot rename an index, so each is recreated.
ALTER TABLE `payment_policies` RENAME TO `payment_receive_policies`;
DROP INDEX `idx_payment_policies_agent_id`;
CREATE UNIQUE INDEX `idx_payment_receive_policies_agent_id` ON `payment_receive_policies` (`agent_id`);

ALTER TABLE `payment_rules` RENAME TO `payment_receive_rules`;
DROP INDEX `idx_payment_rules_policy_id`;
CREATE INDEX `idx_payment_receive_rules_policy_id` ON `payment_receive_rules` (`policy_id`);

ALTER TABLE `payment_rule_x402_configs` RENAME TO `payment_receive_rule_x402_configs`;
ALTER TABLE `payment_receive_rule_x402_configs` RENAME COLUMN `payment_rule_id` TO `payment_receive_rule_id`;

-- the per-agent enabled flag is keyed by tool name
UPDATE "self_tools" SET name = 'create_payment_receive_rule' WHERE name = 'create_payment_rule';
UPDATE "self_tools" SET name = 'update_payment_receive_rule' WHERE name = 'update_payment_rule';
UPDATE "self_tools" SET name = 'delete_payment_receive_rule' WHERE name = 'delete_payment_rule';
UPDATE "self_tools" SET name = 'list_payment_receive_rules' WHERE name = 'list_payment_rules';
UPDATE "self_tools" SET name = 'get_payment_receive_policy' WHERE name = 'get_payment_policy';
UPDATE "self_tools" SET name = 'delete_payment_receive_policy' WHERE name = 'delete_payment_policy';

-- +goose Down
ALTER TABLE `payment_receive_rule_x402_configs` RENAME COLUMN `payment_receive_rule_id` TO `payment_rule_id`;
ALTER TABLE `payment_receive_rule_x402_configs` RENAME TO `payment_rule_x402_configs`;

DROP INDEX `idx_payment_receive_rules_policy_id`;
ALTER TABLE `payment_receive_rules` RENAME TO `payment_rules`;
CREATE INDEX `idx_payment_rules_policy_id` ON `payment_rules` (`policy_id`);

DROP INDEX `idx_payment_receive_policies_agent_id`;
ALTER TABLE `payment_receive_policies` RENAME TO `payment_policies`;
CREATE UNIQUE INDEX `idx_payment_policies_agent_id` ON `payment_policies` (`agent_id`);

UPDATE "self_tools" SET name = 'create_payment_rule' WHERE name = 'create_payment_receive_rule';
UPDATE "self_tools" SET name = 'update_payment_rule' WHERE name = 'update_payment_receive_rule';
UPDATE "self_tools" SET name = 'delete_payment_rule' WHERE name = 'delete_payment_receive_rule';
UPDATE "self_tools" SET name = 'list_payment_rules' WHERE name = 'list_payment_receive_rules';
UPDATE "self_tools" SET name = 'get_payment_policy' WHERE name = 'get_payment_receive_policy';
UPDATE "self_tools" SET name = 'delete_payment_policy' WHERE name = 'delete_payment_receive_policy';
