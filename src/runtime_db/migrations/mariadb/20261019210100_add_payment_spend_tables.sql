-- +goose Up
-- create "payment_spend_policies" table
CREATE TABLE `payment_spend_policies` (
  `id` varchar(36) NOT NULL,
  `agent_id` varchar(36) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_payment_spend_policies_agent_id` (`agent_id`),
  CONSTRAINT `fk_payment_spend_policies_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- create "payment_spend_rules" table
CREATE TABLE `payment_spend_rules` (
  `id` varchar(36) NOT NULL,
  `policy_id` varchar(36) NOT NULL,
  `position` bigint NOT NULL,
  `name` varchar(64) NOT NULL,
  `max_amount_per_payment` varchar(78) NOT NULL,
  `method` varchar(32) NOT NULL,
  `window_seconds` bigint NOT NULL DEFAULT 0,
  `max_amount_per_window` varchar(78) NOT NULL DEFAULT '0',
  `window_spent` varchar(78) NOT NULL DEFAULT '0',
  `window_started_at` datetime(3) NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_payment_spend_rules_policy_id` (`policy_id`),
  CONSTRAINT `fk_payment_spend_rules_policy` FOREIGN KEY (`policy_id`) REFERENCES `payment_spend_policies` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- create "payment_spend_rule_x402_configs" table
CREATE TABLE `payment_spend_rule_x402_configs` (
  `payment_spend_rule_id` varchar(36) NOT NULL,
  `network` varchar(64) NOT NULL,
  `asset` varchar(42) NOT NULL,
  `asset_transfer_method` varchar(16) NULL,
  `pay_to_patterns` longtext NOT NULL,
  PRIMARY KEY (`payment_spend_rule_id`),
  CONSTRAINT `fk_payment_spend_rule_x402_configs_payment_spend_rule` FOREIGN KEY (`payment_spend_rule_id`) REFERENCES `payment_spend_rules` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;

-- +goose Down
-- reverse: create "payment_spend_rule_x402_configs" table
DROP TABLE `payment_spend_rule_x402_configs`;
-- reverse: create "payment_spend_rules" table
DROP TABLE `payment_spend_rules`;
-- reverse: create "payment_spend_policies" table
DROP TABLE `payment_spend_policies`;
