-- +goose Up
-- create "conversation_messages" table
CREATE TABLE `conversation_messages` (
  `id` varchar(191) NOT NULL,
  `conversation_id` longtext NULL,
  `author` longtext NULL,
  `role` longtext NULL,
  `platform` longtext NULL,
  `content` longtext NULL,
  `hash` longtext NULL,
  `timestamp` datetime(3) NULL,
  `created_at` datetime(3) NULL,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- create "llm_connections" table
CREATE TABLE `llm_connections` (
  `id` varchar(191) NOT NULL,
  `platform` longtext NULL,
  `model` longtext NULL,
  `base_url` longtext NULL,
  `setup_status` longtext NULL,
  `created_at` datetime(3) NULL,
  `updated_at` datetime(3) NULL,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- create "mcp_connections" table
CREATE TABLE `mcp_connections` (
  `id` varchar(191) NOT NULL,
  `name` longtext NULL,
  `url` longtext NULL,
  `header_auth_keys` json NULL,
  `created_at` datetime(3) NULL,
  `updated_at` datetime(3) NULL,
  PRIMARY KEY (`id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;

-- +goose Down
-- reverse: create "mcp_connections" table
DROP TABLE `mcp_connections`;
-- reverse: create "llm_connections" table
DROP TABLE `llm_connections`;
-- reverse: create "conversation_messages" table
DROP TABLE `conversation_messages`;
