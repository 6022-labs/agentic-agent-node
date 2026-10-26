-- +goose Up
-- disable the enforcement of foreign-keys constraints
PRAGMA foreign_keys = off;
-- create "payment_spend_policies" table
CREATE TABLE `payment_spend_policies` (
  `id` text NULL,
  `agent_id` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_payment_spend_policies_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_payment_spend_policies_agent_id" to table: "payment_spend_policies"
CREATE UNIQUE INDEX `idx_payment_spend_policies_agent_id` ON `payment_spend_policies` (`agent_id`);
-- create "payment_spend_rules" table
CREATE TABLE `payment_spend_rules` (
  `id` text NULL,
  `policy_id` text NOT NULL,
  `position` integer NOT NULL,
  `name` text NOT NULL,
  `max_amount` text NOT NULL,
  `method` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_payment_spend_rules_policy` FOREIGN KEY (`policy_id`) REFERENCES `payment_spend_policies` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_payment_spend_rules_policy_id" to table: "payment_spend_rules"
CREATE INDEX `idx_payment_spend_rules_policy_id` ON `payment_spend_rules` (`policy_id`);
-- create "payment_spend_rule_x402_configs" table
CREATE TABLE `payment_spend_rule_x402_configs` (
  `payment_spend_rule_id` text NULL,
  `network_patterns` text NOT NULL,
  `asset_patterns` text NOT NULL,
  `pay_to_patterns` text NOT NULL,
  `asset_transfer_method_patterns` text NOT NULL,
  PRIMARY KEY (`payment_spend_rule_id`),
  CONSTRAINT `fk_payment_spend_rule_x402_configs_payment_spend_rule` FOREIGN KEY (`payment_spend_rule_id`) REFERENCES `payment_spend_rules` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- enable back the enforcement of foreign-keys constraints
PRAGMA foreign_keys = on;

-- +goose Down
-- reverse: create "payment_spend_rule_x402_configs" table
DROP TABLE `payment_spend_rule_x402_configs`;
-- reverse: create index "idx_payment_spend_rules_policy_id" to table: "payment_spend_rules"
DROP INDEX `idx_payment_spend_rules_policy_id`;
-- reverse: create "payment_spend_rules" table
DROP TABLE `payment_spend_rules`;
-- reverse: create index "idx_payment_spend_policies_agent_id" to table: "payment_spend_policies"
DROP INDEX `idx_payment_spend_policies_agent_id`;
-- reverse: create "payment_spend_policies" table
DROP TABLE `payment_spend_policies`;
