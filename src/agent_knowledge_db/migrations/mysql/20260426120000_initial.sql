-- +goose Up
-- create "conversation_messages" table — local agent transcript per conversation,
-- deduped by SHA-512 hash of (role, author, content, platform, conversationId).
CREATE TABLE `conversation_messages` (
  `id`              char(36)        NOT NULL,
  `conversation_id` varchar(128)    NOT NULL,
  `author`          varchar(128)    NOT NULL,
  `role`            varchar(32)     NOT NULL,
  `platform`        varchar(64)     NULL,
  `content`         text            NOT NULL,
  `hash`            varchar(128)    NOT NULL,
  `timestamp`       datetime(3)     NOT NULL,
  `created_at`      datetime(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uq_conversation_messages_hash` (`hash`),
  INDEX `idx_conversation_messages_conversation_id` (`conversation_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;

-- +goose Down
DROP TABLE IF EXISTS `conversation_messages`;
