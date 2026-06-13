-- +goose Up
-- create "payment_policies" table
CREATE TABLE `payment_policies` (
  `id` varchar(36) NOT NULL,
  `agent_id` varchar(36) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_payment_policies_agent_id` (`agent_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- create "payment_rules" table
CREATE TABLE `payment_rules` (
  `id` varchar(36) NOT NULL,
  `policy_id` varchar(36) NOT NULL,
  `position` bigint NOT NULL,
  `name` varchar(64) NOT NULL,
  `network` varchar(64) NOT NULL,
  `asset` varchar(42) NOT NULL,
  `pay_to` varchar(42) NOT NULL,
  `price` varchar(78) NOT NULL,
  `max_timeout_seconds` bigint NOT NULL,
  `eip712_name` varchar(64) NOT NULL,
  `eip712_version` varchar(16) NOT NULL,
  `address_patterns` longtext NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_payment_rules_policy_id` (`policy_id`),
  CONSTRAINT `fk_payment_rules_policy` FOREIGN KEY (`policy_id`) REFERENCES `payment_policies` (`id`) ON UPDATE RESTRICT ON DELETE RESTRICT
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;

-- +goose Down
-- reverse: create "payment_rules" table
DROP TABLE `payment_rules`;
-- reverse: create "payment_policies" table
DROP TABLE `payment_policies`;
