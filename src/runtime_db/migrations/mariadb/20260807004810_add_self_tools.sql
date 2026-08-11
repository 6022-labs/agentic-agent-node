-- +goose Up
-- create "self_tools" table
CREATE TABLE `self_tools` (
  `id` varchar(36) NOT NULL,
  `agent_id` varchar(36) NOT NULL,
  `name` varchar(128) NOT NULL,
  `enabled` bool NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uq_self_tools_agent_name` (`agent_id`, `name`),
  CONSTRAINT `fk_self_tools_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;

-- +goose Down
-- reverse: create "self_tools" table
DROP TABLE `self_tools`;
