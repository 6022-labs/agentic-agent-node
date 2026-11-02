-- +goose Up
-- One rule per token: old rules become price rows under the token's first rule; equal prices merge.
CREATE TABLE `payment_receive_rule_prices` (
  `id` varchar(36) NOT NULL,
  `payment_receive_rule_id` varchar(36) NOT NULL,
  `price` varchar(78) NOT NULL,
  `address_patterns` longtext NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uq_payment_receive_rule_prices_rule_price` (`payment_receive_rule_id`, `price`),
  CONSTRAINT `fk_payment_receive_rule_prices_payment_receive_rule` FOREIGN KEY (`payment_receive_rule_id`) REFERENCES `payment_receive_rules` (`id`) ON UPDATE RESTRICT ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;

ALTER TABLE `payment_receive_rule_x402_configs` ADD COLUMN `policy_id` varchar(36) NULL AFTER `payment_receive_rule_id`;
-- canonicalize like the API so equal values group
UPDATE `payment_receive_rules`
SET `price` = IF(TRIM(LEADING '0' FROM `price`) = '', '0', TRIM(LEADING '0' FROM `price`));
UPDATE `payment_receive_rule_x402_configs` AS c
JOIN `payment_receive_rules` AS r ON r.`id` = c.`payment_receive_rule_id`
SET c.`policy_id` = r.`policy_id`,
    c.`asset` = LOWER(c.`asset`),
    c.`network` = IF(
      c.`network` LIKE 'eip155:%' AND TRIM(LEADING '0' FROM SUBSTRING(c.`network`, 8)) <> '',
      CONCAT('eip155:', TRIM(LEADING '0' FROM SUBSTRING(c.`network`, 8))),
      c.`network`
    );

-- GROUP_CONCAT truncates at 1024 bytes by default.
SET SESSION group_concat_max_len = 16777216;
INSERT INTO `payment_receive_rule_prices` (`id`, `payment_receive_rule_id`, `price`, `address_patterns`, `created_at`, `updated_at`)
SELECT UUID(), g.`keeper_id`, g.`price`, CONCAT('[', GROUP_CONCAT(JSON_QUOTE(g.`pattern`) ORDER BY g.`sort` SEPARATOR ','), ']'), MIN(g.`created_at`), NOW(3)
FROM (
  SELECT k.`keeper_id`, k.`price`, j.`pattern`, MIN(k.`rnk` * 100000 + j.`ord`) AS `sort`, MIN(k.`created_at`) AS `created_at`
  FROM (
    SELECT r.`id`, r.`price`, r.`created_at`, c.`address_patterns`,
           ROW_NUMBER() OVER w AS `rnk`,
           FIRST_VALUE(r.`id`) OVER w AS `keeper_id`
    FROM `payment_receive_rules` AS r
    JOIN `payment_receive_rule_x402_configs` AS c ON c.`payment_receive_rule_id` = r.`id`
    WINDOW w AS (PARTITION BY c.`policy_id`, c.`network`, c.`asset` ORDER BY r.`position`, r.`created_at`, r.`id`)
  ) AS k,
  JSON_TABLE(k.`address_patterns`, '$[*]' COLUMNS (`ord` FOR ORDINALITY, `pattern` varchar(4096) PATH '$')) AS j
  GROUP BY k.`keeper_id`, k.`price`, j.`pattern`
) AS g
GROUP BY g.`keeper_id`, g.`price`;

DELETE r FROM `payment_receive_rules` AS r
JOIN (
  SELECT ranked.`id`
  FROM (
    SELECT r2.`id`,
           ROW_NUMBER() OVER (PARTITION BY c2.`policy_id`, c2.`network`, c2.`asset` ORDER BY r2.`position`, r2.`created_at`, r2.`id`) AS `rnk`
    FROM `payment_receive_rules` AS r2
    JOIN `payment_receive_rule_x402_configs` AS c2 ON c2.`payment_receive_rule_id` = r2.`id`
  ) AS ranked
  WHERE ranked.`rnk` > 1
) AS doomed ON doomed.`id` = r.`id`;

ALTER TABLE `payment_receive_rule_x402_configs` DROP COLUMN `address_patterns`, MODIFY COLUMN `policy_id` varchar(36) NOT NULL, ADD UNIQUE INDEX `uq_payment_receive_rule_x402_configs_policy_asset` (`policy_id`, `network`, `asset`);
ALTER TABLE `payment_receive_rules` DROP COLUMN `position`, DROP COLUMN `price`;

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
