-- +goose Up
-- create "agents" table
CREATE TABLE `agents` (
  `id` varchar(36) NOT NULL,
  `chain_id` varchar(64) NOT NULL,
  `collection_token_id` bigint unsigned NULL,
  `collection_address` varchar(42) NULL,
  `proposal_id` bigint unsigned NULL,
  `name` varchar(255) NOT NULL,
  `clone_of` bigint unsigned NULL,
  `ens_domain` varchar(255) NOT NULL,
  `role` varchar(32) NOT NULL,
  `creator` varchar(42) NOT NULL,
  `wallet_address` varchar(42) NOT NULL,
  `owner` varchar(42) NULL,
  `images` longtext NULL,
  `attributes` longtext NULL,
  `updated_at` datetime(3) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- create "agent_registration_states" table
CREATE TABLE `agent_registration_states` (
  `id` varchar(36) NOT NULL,
  `agent_id` varchar(36) NOT NULL,
  `mint_status` varchar(32) NOT NULL,
  `mint_chain_id` bigint unsigned NULL,
  `mint_tx_hash` varchar(66) NULL,
  `minted_at` datetime(3) NULL,
  `ens_registry_status` varchar(32) NOT NULL,
  `ens_chain_id` bigint unsigned NULL,
  `ens_tx_hash` varchar(66) NULL,
  `ens_published_at` datetime(3) NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_agent_registration_states_agent_id` (`agent_id`),
  CONSTRAINT `fk_agent_registration_states_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- create "conversations" table
CREATE TABLE `conversations` (
  `id` varchar(128) NOT NULL,
  `agent_id` varchar(36) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_conversations_agent_id` (`agent_id`),
  CONSTRAINT `fk_conversations_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- create "conversation_messages" table
CREATE TABLE `conversation_messages` (
  `id` varchar(36) NOT NULL,
  `conversation_id` varchar(128) NOT NULL,
  `role` varchar(32) NOT NULL,
  `content` text NULL,
  `tool_calls` text NULL,
  `tool_call_id` varchar(128) NULL,
  `hash` varchar(128) NOT NULL,
  `timestamp` datetime(3) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_conversation_messages_conversation_id` (`conversation_id`),
  UNIQUE INDEX `uq_conversation_messages_hash` (`hash`),
  CONSTRAINT `fk_conversation_messages_conversation` FOREIGN KEY (`conversation_id`) REFERENCES `conversations` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- create "llm_connections" table
CREATE TABLE `llm_connections` (
  `id` varchar(36) NOT NULL,
  `agent_id` varchar(36) NOT NULL,
  `platform` varchar(64) NOT NULL,
  `model` varchar(255) NOT NULL,
  `base_url` text NULL,
  `setup_status` varchar(32) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_llm_connections_agent_id` (`agent_id`),
  CONSTRAINT `fk_llm_connections_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- create "mcp_connections" table
CREATE TABLE `mcp_connections` (
  `id` varchar(36) NOT NULL,
  `agent_id` varchar(36) NOT NULL,
  `name` varchar(128) NOT NULL,
  `url` text NOT NULL,
  `header_auth_keys` json NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_mcp_connections_agent_id` (`agent_id`),
  CONSTRAINT `fk_mcp_connections_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- create "payment_policies" table
CREATE TABLE `payment_policies` (
  `id` varchar(36) NOT NULL,
  `agent_id` varchar(36) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_payment_policies_agent_id` (`agent_id`),
  CONSTRAINT `fk_payment_policies_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- create "payment_rules" table
CREATE TABLE `payment_rules` (
  `id` varchar(36) NOT NULL,
  `policy_id` varchar(36) NOT NULL,
  `position` bigint NOT NULL,
  `name` varchar(64) NOT NULL,
  `price` varchar(78) NOT NULL,
  `method` varchar(32) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_payment_rules_policy_id` (`policy_id`),
  CONSTRAINT `fk_payment_rules_policy` FOREIGN KEY (`policy_id`) REFERENCES `payment_policies` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- create "payment_rule_x402_configs" table
CREATE TABLE `payment_rule_x402_configs` (
  `payment_rule_id` varchar(36) NOT NULL,
  `network` varchar(64) NOT NULL,
  `asset` varchar(42) NOT NULL,
  `pay_to` varchar(42) NOT NULL,
  `asset_transfer_method` varchar(16) NOT NULL,
  `eip712_name` varchar(64) NULL,
  `eip712_version` varchar(16) NULL,
  `address_patterns` longtext NOT NULL,
  PRIMARY KEY (`payment_rule_id`),
  CONSTRAINT `fk_payment_rule_x402_configs_payment_rule` FOREIGN KEY (`payment_rule_id`) REFERENCES `payment_rules` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;

-- +goose Down
-- reverse: create "payment_rule_x402_configs" table
DROP TABLE `payment_rule_x402_configs`;
-- reverse: create "payment_rules" table
DROP TABLE `payment_rules`;
-- reverse: create "payment_policies" table
DROP TABLE `payment_policies`;
-- reverse: create "mcp_connections" table
DROP TABLE `mcp_connections`;
-- reverse: create "llm_connections" table
DROP TABLE `llm_connections`;
-- reverse: create "conversation_messages" table
DROP TABLE `conversation_messages`;
-- reverse: create "conversations" table
DROP TABLE `conversations`;
-- reverse: create "agent_registration_states" table
DROP TABLE `agent_registration_states`;
-- reverse: create "agents" table
DROP TABLE `agents`;
