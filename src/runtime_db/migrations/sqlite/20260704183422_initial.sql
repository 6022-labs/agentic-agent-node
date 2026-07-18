-- +goose Up
-- create "agents" table
CREATE TABLE `agents` (
  `id` text NULL,
  `chain_id` text NULL,
  `collection_token_id` integer NULL,
  `collection_address` text NULL,
  `proposal_id` integer NULL,
  `name` text NULL,
  `clone_of` integer NULL,
  `ens_domain` text NULL,
  `role` text NULL,
  `creator` text NULL,
  `owner` text NULL,
  `soul` text NULL,
  `images` text NULL,
  `attributes` text NULL,
  `updated_at` datetime NOT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`)
);
-- create "mcp_connections" table
CREATE TABLE `mcp_connections` (
  `id` text NULL,
  `agent_id` text NOT NULL,
  `name` text NOT NULL,
  `url` text NOT NULL,
  `header_auth_keys` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_mcp_connections_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "uq_mcp_connections_agent_name" to table: "mcp_connections"
CREATE UNIQUE INDEX `uq_mcp_connections_agent_name` ON `mcp_connections` (`agent_id`, `name`);
-- create "llm_connections" table
CREATE TABLE `llm_connections` (
  `id` text NULL,
  `agent_id` text NOT NULL,
  `platform` text NOT NULL,
  `model` text NOT NULL,
  `base_url` text NULL,
  `setup_status` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_llm_connections_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_llm_connections_agent_id" to table: "llm_connections"
CREATE UNIQUE INDEX `idx_llm_connections_agent_id` ON `llm_connections` (`agent_id`);
-- create "agent_guidances" table
CREATE TABLE `agent_guidances` (
  `id` text NULL,
  `agent_id` text NOT NULL,
  `content` text NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_agent_guidances_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_agent_guidances_agent_id" to table: "agent_guidances"
CREATE UNIQUE INDEX `idx_agent_guidances_agent_id` ON `agent_guidances` (`agent_id`);
-- create "conversations" table
CREATE TABLE `conversations` (
  `id` text NULL,
  `agent_id` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_conversations_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_conversations_agent_id" to table: "conversations"
CREATE INDEX `idx_conversations_agent_id` ON `conversations` (`agent_id`);
-- create "conversation_messages" table
CREATE TABLE `conversation_messages` (
  `id` text NULL,
  `conversation_id` text NOT NULL,
  `role` text NOT NULL,
  `content` text NULL,
  `tool_calls` text NULL,
  `tool_call_id` text NULL,
  `hash` text NOT NULL,
  `timestamp` datetime NOT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_conversation_messages_conversation` FOREIGN KEY (`conversation_id`) REFERENCES `conversations` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "uq_conversation_messages_hash" to table: "conversation_messages"
CREATE UNIQUE INDEX `uq_conversation_messages_hash` ON `conversation_messages` (`hash`);
-- create index "idx_conversation_messages_conversation_id" to table: "conversation_messages"
CREATE INDEX `idx_conversation_messages_conversation_id` ON `conversation_messages` (`conversation_id`);
-- create "agent_registration_states" table
CREATE TABLE `agent_registration_states` (
  `id` text NULL,
  `agent_id` text NOT NULL,
  `mint_status` text NOT NULL,
  `mint_chain_id` integer NULL,
  `mint_tx_hash` text NULL,
  `minted_at` datetime NULL,
  `ens_registry_status` text NOT NULL,
  `ens_chain_id` integer NULL,
  `ens_tx_hash` text NULL,
  `ens_published_at` datetime NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_agent_registration_states_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_agent_registration_states_agent_id" to table: "agent_registration_states"
CREATE UNIQUE INDEX `idx_agent_registration_states_agent_id` ON `agent_registration_states` (`agent_id`);
-- create "agent_addresses" table
CREATE TABLE `agent_addresses` (
  `agent_id` text NULL,
  `type` text NULL,
  `address` text NOT NULL,
  `updated_at` datetime NOT NULL,
  `created_at` datetime NOT NULL,
  PRIMARY KEY (`agent_id`, `type`)
);
-- create "payment_policies" table
CREATE TABLE `payment_policies` (
  `id` text NULL,
  `agent_id` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_payment_policies_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_payment_policies_agent_id" to table: "payment_policies"
CREATE UNIQUE INDEX `idx_payment_policies_agent_id` ON `payment_policies` (`agent_id`);
-- create "payment_rules" table
CREATE TABLE `payment_rules` (
  `id` text NULL,
  `policy_id` text NOT NULL,
  `position` integer NOT NULL,
  `name` text NOT NULL,
  `price` text NOT NULL,
  `method` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_payment_rules_policy` FOREIGN KEY (`policy_id`) REFERENCES `payment_policies` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_payment_rules_policy_id" to table: "payment_rules"
CREATE INDEX `idx_payment_rules_policy_id` ON `payment_rules` (`policy_id`);
-- create "payment_rule_x402_configs" table
CREATE TABLE `payment_rule_x402_configs` (
  `payment_rule_id` text NULL,
  `network` text NOT NULL,
  `asset` text NOT NULL,
  `pay_to` text NOT NULL,
  `asset_transfer_method` text NOT NULL,
  `eip712_name` text NULL,
  `eip712_version` text NULL,
  `address_patterns` text NOT NULL,
  PRIMARY KEY (`payment_rule_id`),
  CONSTRAINT `fk_payment_rule_x402_configs_payment_rule` FOREIGN KEY (`payment_rule_id`) REFERENCES `payment_rules` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "payment_rule_x402_configs" table
DROP TABLE `payment_rule_x402_configs`;
-- reverse: create index "idx_payment_rules_policy_id" to table: "payment_rules"
DROP INDEX `idx_payment_rules_policy_id`;
-- reverse: create "payment_rules" table
DROP TABLE `payment_rules`;
-- reverse: create index "idx_payment_policies_agent_id" to table: "payment_policies"
DROP INDEX `idx_payment_policies_agent_id`;
-- reverse: create "payment_policies" table
DROP TABLE `payment_policies`;
-- reverse: create "agent_addresses" table
DROP TABLE `agent_addresses`;
-- reverse: create index "idx_agent_registration_states_agent_id" to table: "agent_registration_states"
DROP INDEX `idx_agent_registration_states_agent_id`;
-- reverse: create "agent_registration_states" table
DROP TABLE `agent_registration_states`;
-- reverse: create index "idx_conversation_messages_conversation_id" to table: "conversation_messages"
DROP INDEX `idx_conversation_messages_conversation_id`;
-- reverse: create index "uq_conversation_messages_hash" to table: "conversation_messages"
DROP INDEX `uq_conversation_messages_hash`;
-- reverse: create "conversation_messages" table
DROP TABLE `conversation_messages`;
-- reverse: create index "idx_conversations_agent_id" to table: "conversations"
DROP INDEX `idx_conversations_agent_id`;
-- reverse: create "conversations" table
DROP TABLE `conversations`;
-- reverse: create index "idx_agent_guidances_agent_id" to table: "agent_guidances"
DROP INDEX `idx_agent_guidances_agent_id`;
-- reverse: create "agent_guidances" table
DROP TABLE `agent_guidances`;
-- reverse: create index "idx_llm_connections_agent_id" to table: "llm_connections"
DROP INDEX `idx_llm_connections_agent_id`;
-- reverse: create "llm_connections" table
DROP TABLE `llm_connections`;
-- reverse: create index "uq_mcp_connections_agent_name" to table: "mcp_connections"
DROP INDEX `uq_mcp_connections_agent_name`;
-- reverse: create "mcp_connections" table
DROP TABLE `mcp_connections`;
-- reverse: create "agents" table
DROP TABLE `agents`;
