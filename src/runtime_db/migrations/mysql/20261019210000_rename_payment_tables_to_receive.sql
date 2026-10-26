-- +goose Up
-- Renames only, hand-written: atlas diffs a rename as DROP+CREATE, which would
-- destroy every deployed agent's pricing rules. MySQL has no RENAME CONSTRAINT,
-- so each foreign key is dropped and re-added under its new name.
RENAME TABLE `payment_policies` TO `payment_receive_policies`;
ALTER TABLE `payment_receive_policies` RENAME INDEX `idx_payment_policies_agent_id` TO `idx_payment_receive_policies_agent_id`;
ALTER TABLE `payment_receive_policies` DROP FOREIGN KEY `fk_payment_policies_agent`;
ALTER TABLE `payment_receive_policies` ADD CONSTRAINT `fk_payment_receive_policies_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE;

RENAME TABLE `payment_rules` TO `payment_receive_rules`;
ALTER TABLE `payment_receive_rules` RENAME INDEX `idx_payment_rules_policy_id` TO `idx_payment_receive_rules_policy_id`;
ALTER TABLE `payment_receive_rules` DROP FOREIGN KEY `fk_payment_rules_policy`;
ALTER TABLE `payment_receive_rules` ADD CONSTRAINT `fk_payment_receive_rules_policy` FOREIGN KEY (`policy_id`) REFERENCES `payment_receive_policies` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION;

RENAME TABLE `payment_rule_x402_configs` TO `payment_receive_rule_x402_configs`;
ALTER TABLE `payment_receive_rule_x402_configs` DROP FOREIGN KEY `fk_payment_rule_x402_configs_payment_rule`;
ALTER TABLE `payment_receive_rule_x402_configs` RENAME COLUMN `payment_rule_id` TO `payment_receive_rule_id`;
ALTER TABLE `payment_receive_rule_x402_configs` ADD CONSTRAINT `fk_payment_receive_rule_x402_configs_payment_receive_rule` FOREIGN KEY (`payment_receive_rule_id`) REFERENCES `payment_receive_rules` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE;

-- the per-agent enabled flag is keyed by tool name
UPDATE `self_tools` SET name = 'create_payment_receive_rule' WHERE name = 'create_payment_rule';
UPDATE `self_tools` SET name = 'update_payment_receive_rule' WHERE name = 'update_payment_rule';
UPDATE `self_tools` SET name = 'delete_payment_receive_rule' WHERE name = 'delete_payment_rule';
UPDATE `self_tools` SET name = 'list_payment_receive_rules' WHERE name = 'list_payment_rules';
UPDATE `self_tools` SET name = 'get_payment_receive_policy' WHERE name = 'get_payment_policy';
UPDATE `self_tools` SET name = 'delete_payment_receive_policy' WHERE name = 'delete_payment_policy';

-- +goose Down
ALTER TABLE `payment_receive_rule_x402_configs` DROP FOREIGN KEY `fk_payment_receive_rule_x402_configs_payment_receive_rule`;
ALTER TABLE `payment_receive_rule_x402_configs` RENAME COLUMN `payment_receive_rule_id` TO `payment_rule_id`;
RENAME TABLE `payment_receive_rule_x402_configs` TO `payment_rule_x402_configs`;
ALTER TABLE `payment_rule_x402_configs` ADD CONSTRAINT `fk_payment_rule_x402_configs_payment_rule` FOREIGN KEY (`payment_rule_id`) REFERENCES `payment_rules` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE;

ALTER TABLE `payment_receive_rules` DROP FOREIGN KEY `fk_payment_receive_rules_policy`;
ALTER TABLE `payment_receive_rules` RENAME INDEX `idx_payment_receive_rules_policy_id` TO `idx_payment_rules_policy_id`;
RENAME TABLE `payment_receive_rules` TO `payment_rules`;

ALTER TABLE `payment_receive_policies` DROP FOREIGN KEY `fk_payment_receive_policies_agent`;
ALTER TABLE `payment_receive_policies` RENAME INDEX `idx_payment_receive_policies_agent_id` TO `idx_payment_policies_agent_id`;
RENAME TABLE `payment_receive_policies` TO `payment_policies`;
ALTER TABLE `payment_policies` ADD CONSTRAINT `fk_payment_policies_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE;
ALTER TABLE `payment_rules` ADD CONSTRAINT `fk_payment_rules_policy` FOREIGN KEY (`policy_id`) REFERENCES `payment_policies` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION;

UPDATE `self_tools` SET name = 'create_payment_rule' WHERE name = 'create_payment_receive_rule';
UPDATE `self_tools` SET name = 'update_payment_rule' WHERE name = 'update_payment_receive_rule';
UPDATE `self_tools` SET name = 'delete_payment_rule' WHERE name = 'delete_payment_receive_rule';
UPDATE `self_tools` SET name = 'list_payment_rules' WHERE name = 'list_payment_receive_rules';
UPDATE `self_tools` SET name = 'get_payment_policy' WHERE name = 'get_payment_receive_policy';
UPDATE `self_tools` SET name = 'delete_payment_policy' WHERE name = 'delete_payment_receive_policy';
