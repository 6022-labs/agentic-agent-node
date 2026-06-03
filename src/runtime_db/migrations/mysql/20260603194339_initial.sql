-- +goose Up
-- create "agent_registration_states" table
CREATE TABLE `agent_registration_states` (
  `id` varchar(36) NOT NULL,
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
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- create "agents" table
CREATE TABLE `agents` (
  `id` varchar(36) NOT NULL,
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
-- create "conversation_messages" table
CREATE TABLE `conversation_messages` (
  `id` varchar(36) NOT NULL,
  `conversation_id` varchar(128) NOT NULL,
  `role` varchar(32) NOT NULL,
  `content` text NOT NULL,
  `hash` varchar(128) NOT NULL,
  `timestamp` datetime(3) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  INDEX `idx_conversation_messages_conversation_id` (`conversation_id`),
  UNIQUE INDEX `uq_conversation_messages_hash` (`hash`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- create "llm_connections" table
CREATE TABLE `llm_connections` (
  `id` varchar(36) NOT NULL,
  `platform` varchar(64) NOT NULL,
  `model` varchar(255) NOT NULL,
  `base_url` text NULL,
  `setup_status` varchar(32) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
-- create "mcp_connections" table
CREATE TABLE `mcp_connections` (
  `id` varchar(36) NOT NULL,
  `name` varchar(128) NOT NULL,
  `url` text NOT NULL,
  `header_auth_keys` json NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;

-- +goose Down
-- reverse: create "mcp_connections" table
DROP TABLE `mcp_connections`;
-- reverse: create "llm_connections" table
DROP TABLE `llm_connections`;
-- reverse: create "conversation_messages" table
DROP TABLE `conversation_messages`;
-- reverse: create "agents" table
DROP TABLE `agents`;
-- reverse: create "agent_registration_states" table
DROP TABLE `agent_registration_states`;
