-- +goose Up
-- One rule per token: old rules become price rows under the token's first rule; equal prices merge.
-- no-op inside goose's transaction: foreign keys stay on, so drop order matters
PRAGMA foreign_keys = off;
-- canonicalize like the API so equal values group
UPDATE `payment_receive_rules`
SET `price` = CASE WHEN LTRIM(`price`, '0') = '' THEN '0' ELSE LTRIM(`price`, '0') END
WHERE `price` <> '' AND `price` NOT GLOB '*[^0-9]*';
UPDATE `payment_receive_rule_x402_configs`
SET `asset` = LOWER(`asset`),
    `network` = CASE
      WHEN SUBSTR(`network`, 1, 7) = 'eip155:' AND LTRIM(SUBSTR(`network`, 8), '0') <> '' THEN 'eip155:' || LTRIM(SUBSTR(`network`, 8), '0')
      ELSE `network`
    END;
CREATE TABLE `receive_rule_price_migration` AS
SELECT g.`keeper_id`, g.`price`,
       CASE WHEN MAX(g.`pattern` = '.*') THEN '[".*"]' ELSE json_group_array(g.`pattern` ORDER BY g.`sort`) END AS `address_patterns`,
       MIN(g.`created_at`) AS `created_at`
FROM (
  SELECT k.`keeper_id`, k.`price`, CAST(j.`value` AS TEXT) AS `pattern`, MIN(k.`rnk` * 100000 + j.`key`) AS `sort`, MIN(k.`created_at`) AS `created_at`
  FROM (
    -- empty or null patterns matched nobody, so they yield no price row
    SELECT r.`id`, r.`price`, r.`created_at`,
           CASE WHEN TRIM(c.`address_patterns`, ' ' || char(9, 10, 13)) IN ('', 'null') THEN '[]' ELSE c.`address_patterns` END AS `address_patterns`,
           ROW_NUMBER() OVER w AS `rnk`,
           FIRST_VALUE(r.`id`) OVER w AS `keeper_id`
    FROM `payment_receive_rules` AS r
    JOIN `payment_receive_rule_x402_configs` AS c ON c.`payment_receive_rule_id` = r.`id`
    WINDOW w AS (PARTITION BY r.`policy_id`, c.`network`, c.`asset` ORDER BY r.`position`, r.`created_at`, r.`id`)
  ) AS k, json_each(CASE WHEN json_type(k.`address_patterns`) = 'array' THEN k.`address_patterns` ELSE '[]' END) AS j
  WHERE j.`value` IS NOT NULL
  GROUP BY k.`keeper_id`, k.`price`, CAST(j.`value` AS TEXT)
) AS g
GROUP BY g.`keeper_id`, g.`price`;
-- merged rules' configs first, so deleting the rules cascades nothing
CREATE TABLE `receive_rule_merged_migration` AS
SELECT ranked.`id`
FROM (
  SELECT r.`id`,
         ROW_NUMBER() OVER (PARTITION BY r.`policy_id`, c.`network`, c.`asset` ORDER BY r.`position`, r.`created_at`, r.`id`) AS `rnk`
  FROM `payment_receive_rules` AS r
  JOIN `payment_receive_rule_x402_configs` AS c ON c.`payment_receive_rule_id` = r.`id`
) AS ranked
WHERE ranked.`rnk` > 1;
DELETE FROM `payment_receive_rule_x402_configs` WHERE `payment_receive_rule_id` IN (SELECT `id` FROM `receive_rule_merged_migration`);
DELETE FROM `payment_receive_rules` WHERE `id` IN (SELECT `id` FROM `receive_rule_merged_migration`);
DROP TABLE `receive_rule_merged_migration`;
-- create "new_payment_receive_rules" table
CREATE TABLE `new_payment_receive_rules` (
  `id` text NULL,
  `policy_id` text NOT NULL,
  `name` text NOT NULL,
  `method` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_payment_receive_rules_policy` FOREIGN KEY (`policy_id`) REFERENCES `payment_receive_policies` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- copy rows from old table "payment_receive_rules" to new temporary table "new_payment_receive_rules"
INSERT INTO `new_payment_receive_rules` (`id`, `policy_id`, `name`, `method`, `created_at`, `updated_at`) SELECT `id`, `policy_id`, `name`, `method`, `created_at`, `updated_at` FROM `payment_receive_rules`;
-- stage configs (policy from the rule) so no referenced table is dropped
CREATE TABLE `receive_rule_config_migration` AS
SELECT c.`payment_receive_rule_id`, r.`policy_id`, c.`network`, c.`asset`, c.`pay_to`, c.`asset_transfer_method`, c.`eip712_name`, c.`eip712_version`
FROM `payment_receive_rule_x402_configs` AS c JOIN `payment_receive_rules` AS r ON r.`id` = c.`payment_receive_rule_id`;
-- drop the old tables after copying rows
DROP TABLE `payment_receive_rule_x402_configs`;
DROP TABLE `payment_receive_rules`;
-- rename temporary table "new_payment_receive_rules" to "payment_receive_rules"
ALTER TABLE `new_payment_receive_rules` RENAME TO `payment_receive_rules`;
-- create index "idx_payment_receive_rules_policy_id" to table: "payment_receive_rules"
CREATE INDEX `idx_payment_receive_rules_policy_id` ON `payment_receive_rules` (`policy_id`);
-- create "payment_receive_rule_x402_configs" table
CREATE TABLE `payment_receive_rule_x402_configs` (
  `payment_receive_rule_id` text NULL,
  `policy_id` text NOT NULL,
  `network` text NOT NULL,
  `asset` text NOT NULL,
  `pay_to` text NOT NULL,
  `asset_transfer_method` text NOT NULL,
  `eip712_name` text NULL,
  `eip712_version` text NULL,
  PRIMARY KEY (`payment_receive_rule_id`),
  CONSTRAINT `fk_payment_receive_rule_x402_configs_payment_receive_rule` FOREIGN KEY (`payment_receive_rule_id`) REFERENCES `payment_receive_rules` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "uq_payment_receive_rule_x402_configs_policy_asset" to table: "payment_receive_rule_x402_configs"
CREATE UNIQUE INDEX `uq_payment_receive_rule_x402_configs_policy_asset` ON `payment_receive_rule_x402_configs` (`policy_id`, `network`, `asset`);
INSERT INTO `payment_receive_rule_x402_configs` (`payment_receive_rule_id`, `policy_id`, `network`, `asset`, `pay_to`, `asset_transfer_method`, `eip712_name`, `eip712_version`) SELECT `payment_receive_rule_id`, `policy_id`, `network`, `asset`, `pay_to`, `asset_transfer_method`, `eip712_name`, `eip712_version` FROM `receive_rule_config_migration`;
DROP TABLE `receive_rule_config_migration`;
-- create "payment_receive_rule_prices" table
CREATE TABLE `payment_receive_rule_prices` (
  `id` text NULL,
  `payment_receive_rule_id` text NOT NULL,
  `price` text NOT NULL,
  `address_patterns` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_payment_receive_rule_prices_payment_receive_rule` FOREIGN KEY (`payment_receive_rule_id`) REFERENCES `payment_receive_rules` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "uq_payment_receive_rule_prices_rule_price" to table: "payment_receive_rule_prices"
CREATE UNIQUE INDEX `uq_payment_receive_rule_prices_rule_price` ON `payment_receive_rule_prices` (`payment_receive_rule_id`, `price`);
-- random v4 ids
INSERT INTO `payment_receive_rule_prices` (`id`, `payment_receive_rule_id`, `price`, `address_patterns`, `created_at`, `updated_at`)
SELECT lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))), 2) || '-'
         || substr('89ab', 1 + (abs(random()) % 4), 1) || substr(lower(hex(randomblob(2))), 2) || '-' || lower(hex(randomblob(6))),
       `keeper_id`, `price`, `address_patterns`, `created_at`, CURRENT_TIMESTAMP
FROM `receive_rule_price_migration`;
DROP TABLE `receive_rule_price_migration`;
-- pairs with the no-op pragma above
PRAGMA foreign_keys = on;

-- +goose Down
-- Lossy: each rule keeps only its cheapest price.
-- no-op inside goose's transaction: children are dropped before their parent
PRAGMA foreign_keys = off;
CREATE TABLE `receive_rule_down_migration` AS
SELECT r.`id`, r.`policy_id`, r.`name`, r.`method`, r.`created_at`, r.`updated_at`,
       COALESCE((SELECT p.`price` FROM `payment_receive_rule_prices` AS p WHERE p.`payment_receive_rule_id` = r.`id` ORDER BY LENGTH(p.`price`), p.`price` LIMIT 1), '0') AS `price`,
       COALESCE((SELECT p.`address_patterns` FROM `payment_receive_rule_prices` AS p WHERE p.`payment_receive_rule_id` = r.`id` ORDER BY LENGTH(p.`price`), p.`price` LIMIT 1), '[]') AS `address_patterns`
FROM `payment_receive_rules` AS r;
CREATE TABLE `receive_rule_config_down_migration` AS SELECT * FROM `payment_receive_rule_x402_configs`;
DROP TABLE `payment_receive_rule_prices`;
DROP TABLE `payment_receive_rule_x402_configs`;
DROP TABLE `payment_receive_rules`;
CREATE TABLE `payment_receive_rules` (
  `id` text NULL,
  `policy_id` text NOT NULL,
  `position` integer NOT NULL,
  `name` text NOT NULL,
  `price` text NOT NULL,
  `method` text NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_payment_receive_rules_policy` FOREIGN KEY (`policy_id`) REFERENCES `payment_receive_policies` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX `idx_payment_receive_rules_policy_id` ON `payment_receive_rules` (`policy_id`);
INSERT INTO `payment_receive_rules` (`id`, `policy_id`, `position`, `name`, `price`, `method`, `created_at`, `updated_at`)
SELECT `id`, `policy_id`, ROW_NUMBER() OVER (PARTITION BY `policy_id` ORDER BY `created_at`, `id`) - 1, `name`, `price`, `method`, `created_at`, `updated_at`
FROM `receive_rule_down_migration`;
CREATE TABLE `payment_receive_rule_x402_configs` (
  `payment_receive_rule_id` text NULL,
  `network` text NOT NULL,
  `asset` text NOT NULL,
  `pay_to` text NOT NULL,
  `asset_transfer_method` text NOT NULL,
  `eip712_name` text NULL,
  `eip712_version` text NULL,
  `address_patterns` text NOT NULL,
  PRIMARY KEY (`payment_receive_rule_id`),
  CONSTRAINT `fk_payment_receive_rule_x402_configs_payment_receive_rule` FOREIGN KEY (`payment_receive_rule_id`) REFERENCES `payment_receive_rules` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
);
INSERT INTO `payment_receive_rule_x402_configs` (`payment_receive_rule_id`, `network`, `asset`, `pay_to`, `asset_transfer_method`, `eip712_name`, `eip712_version`, `address_patterns`)
SELECT c.`payment_receive_rule_id`, c.`network`, c.`asset`, c.`pay_to`, c.`asset_transfer_method`, c.`eip712_name`, c.`eip712_version`, d.`address_patterns`
FROM `receive_rule_config_down_migration` AS c JOIN `receive_rule_down_migration` AS d ON d.`id` = c.`payment_receive_rule_id`;
DROP TABLE `receive_rule_config_down_migration`;
DROP TABLE `receive_rule_down_migration`;
PRAGMA foreign_keys = on;
