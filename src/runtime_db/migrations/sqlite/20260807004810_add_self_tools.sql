-- +goose Up
-- create "self_tools" table
CREATE TABLE `self_tools` (
  `id` text NULL,
  `agent_id` text NOT NULL,
  `name` text NOT NULL,
  `enabled` numeric NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_self_tools_agent` FOREIGN KEY (`agent_id`) REFERENCES `agents` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "uq_self_tools_agent_name" to table: "self_tools"
CREATE UNIQUE INDEX `uq_self_tools_agent_name` ON `self_tools` (`agent_id`, `name`);

-- +goose Down
-- reverse: create index "uq_self_tools_agent_name" to table: "self_tools"
DROP INDEX `uq_self_tools_agent_name`;
-- reverse: create "self_tools" table
DROP TABLE `self_tools`;
