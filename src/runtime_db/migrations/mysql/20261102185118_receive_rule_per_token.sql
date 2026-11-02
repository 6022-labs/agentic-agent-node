-- +goose Up
-- One rule per token: old rules become price rows under the token's first rule; equal prices merge.
-- Everything that can fail on data runs before the first DDL, which commits implicitly here.
-- GROUP_CONCAT truncates at 1024 bytes by default.
SET SESSION group_concat_max_len = 16777216;
-- a failed earlier run on this connection leaves its temporary tables behind
DROP TEMPORARY TABLE IF EXISTS `receive_rule_canonical_migration`, `receive_rule_price_migration`, `receive_rule_merged_migration`;
-- canonicalize like the API so equal values group; empty or null patterns matched nobody
CREATE TEMPORARY TABLE `receive_rule_canonical_migration` AS
SELECT r.`id`, r.`policy_id`, r.`position`, r.`created_at`,
       IF(r.`price` REGEXP '^[0-9]+$', IF(TRIM(LEADING '0' FROM r.`price`) = '', '0', TRIM(LEADING '0' FROM r.`price`)), r.`price`) AS `price`,
       LOWER(c.`asset`) AS `asset`,
       IF(
         BINARY LEFT(c.`network`, 7) = BINARY 'eip155:' AND TRIM(LEADING '0' FROM SUBSTRING(c.`network`, 8)) <> '',
         CONCAT('eip155:', TRIM(LEADING '0' FROM SUBSTRING(c.`network`, 8))),
         c.`network`
       ) AS `network`,
       IF(TRIM(c.`address_patterns`) IN ('', 'null'), '[]', c.`address_patterns`) AS `address_patterns`
FROM `payment_receive_rules` AS r
JOIN `payment_receive_rule_x402_configs` AS c ON c.`payment_receive_rule_id` = r.`id`;
CREATE TEMPORARY TABLE `receive_rule_price_migration` AS
SELECT g.`keeper_id`, g.`price`,
       IF(MAX(g.`pattern` = '.*'), '[".*"]', CONCAT('[', GROUP_CONCAT(JSON_QUOTE(g.`pattern`) ORDER BY g.`sort` SEPARATOR ','), ']')) AS `address_patterns`,
       MIN(g.`created_at`) AS `created_at`
FROM (
  SELECT e.`keeper_id`, e.`price`, e.`pattern`, MIN(e.`rnk` * 100000 + e.`ord`) AS `sort`, MIN(e.`created_at`) AS `created_at`
  FROM (
    -- byte-exact patterns, so case variants stay distinct
    SELECT k.`keeper_id`, k.`price`, k.`rnk`, k.`created_at`, j.`ord`, j.`pattern` COLLATE utf8mb4_bin AS `pattern`
    FROM (
      SELECT m.`id`, m.`price`, m.`created_at`, m.`address_patterns`,
             ROW_NUMBER() OVER w AS `rnk`,
             FIRST_VALUE(m.`id`) OVER w AS `keeper_id`
      FROM `receive_rule_canonical_migration` AS m
      WINDOW w AS (PARTITION BY m.`policy_id`, m.`network`, m.`asset` ORDER BY m.`position`, m.`created_at`, m.`id`)
    ) AS k,
    JSON_TABLE(k.`address_patterns`, '$[*]' COLUMNS (`ord` FOR ORDINALITY, `pattern` varchar(4096) CHARACTER SET utf8mb4 PATH '$')) AS j
    WHERE j.`pattern` IS NOT NULL
  ) AS e
  GROUP BY e.`keeper_id`, e.`price`, e.`pattern`
) AS g
GROUP BY g.`keeper_id`, g.`price`;
CREATE TEMPORARY TABLE `receive_rule_merged_migration` AS
SELECT ranked.`id`
FROM (
  SELECT m.`id`,
         ROW_NUMBER() OVER (PARTITION BY m.`policy_id`, m.`network`, m.`asset` ORDER BY m.`position`, m.`created_at`, m.`id`) AS `rnk`
  FROM `receive_rule_canonical_migration` AS m
) AS ranked
WHERE ranked.`rnk` > 1;

CREATE TABLE `payment_receive_rule_prices` (
  `id` varchar(36) NOT NULL,
  `payment_receive_rule_id` varchar(36) NOT NULL,
  `price` varchar(78) NOT NULL,
  `address_patterns` longtext NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uq_payment_receive_rule_prices_rule_price` (`payment_receive_rule_id`, `price`),
  CONSTRAINT `fk_payment_receive_rule_prices_payment_receive_rule` FOREIGN KEY (`payment_receive_rule_id`) REFERENCES `payment_receive_rules` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;

ALTER TABLE `payment_receive_rule_x402_configs` ADD COLUMN `policy_id` varchar(36) NULL AFTER `payment_receive_rule_id`;
-- key predicate keeps sql_safe_updates happy
UPDATE `payment_receive_rule_x402_configs` AS c
JOIN `receive_rule_canonical_migration` AS m ON m.`id` = c.`payment_receive_rule_id`
SET c.`policy_id` = m.`policy_id`, c.`asset` = m.`asset`, c.`network` = m.`network`
WHERE c.`payment_receive_rule_id` <> '';
INSERT INTO `payment_receive_rule_prices` (`id`, `payment_receive_rule_id`, `price`, `address_patterns`, `created_at`, `updated_at`)
SELECT UUID(), `keeper_id`, `price`, `address_patterns`, `created_at`, NOW(3)
FROM `receive_rule_price_migration`;
DELETE r FROM `payment_receive_rules` AS r
JOIN `receive_rule_merged_migration` AS d ON d.`id` = r.`id`
WHERE r.`id` <> '';

ALTER TABLE `payment_receive_rule_x402_configs` DROP COLUMN `address_patterns`, MODIFY COLUMN `policy_id` varchar(36) NOT NULL, ADD UNIQUE INDEX `uq_payment_receive_rule_x402_configs_policy_asset` (`policy_id`, `network`, `asset`);
ALTER TABLE `payment_receive_rules` DROP COLUMN `position`, DROP COLUMN `price`;
DROP TEMPORARY TABLE `receive_rule_canonical_migration`, `receive_rule_price_migration`, `receive_rule_merged_migration`;

-- +goose Down
-- Lossy: each rule keeps only its cheapest price.
ALTER TABLE `payment_receive_rules` ADD COLUMN `price` varchar(78) NULL, ADD COLUMN `position` bigint NOT NULL DEFAULT 0;
ALTER TABLE `payment_receive_rule_x402_configs` ADD COLUMN `address_patterns` longtext NULL;
UPDATE `payment_receive_rules` AS r
SET r.`price` = COALESCE((
  SELECT p.`price` FROM `payment_receive_rule_prices` AS p
  WHERE p.`payment_receive_rule_id` = r.`id`
  ORDER BY LENGTH(p.`price`), p.`price` LIMIT 1
), '0');
UPDATE `payment_receive_rules` AS r
JOIN (
  SELECT `id`, ROW_NUMBER() OVER (PARTITION BY `policy_id` ORDER BY `created_at`, `id`) - 1 AS `position`
  FROM `payment_receive_rules`
) AS o ON o.`id` = r.`id`
SET r.`position` = o.`position`;
UPDATE `payment_receive_rule_x402_configs` AS c
SET c.`address_patterns` = COALESCE((
  SELECT p.`address_patterns` FROM `payment_receive_rule_prices` AS p
  WHERE p.`payment_receive_rule_id` = c.`payment_receive_rule_id`
  ORDER BY LENGTH(p.`price`), p.`price` LIMIT 1
), '[]');
ALTER TABLE `payment_receive_rules` MODIFY COLUMN `price` varchar(78) NOT NULL, ALTER COLUMN `position` DROP DEFAULT;
ALTER TABLE `payment_receive_rule_x402_configs` DROP INDEX `uq_payment_receive_rule_x402_configs_policy_asset`, DROP COLUMN `policy_id`, MODIFY COLUMN `address_patterns` longtext NOT NULL;
DROP TABLE `payment_receive_rule_prices`;
