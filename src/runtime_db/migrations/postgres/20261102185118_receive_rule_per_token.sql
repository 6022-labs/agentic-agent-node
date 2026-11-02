-- +goose Up
-- One rule per token: old rules become price rows under the token's first rule; equal prices merge.
CREATE TABLE "public"."payment_receive_rule_prices" (
  "id" character varying(36) NOT NULL,
  "payment_receive_rule_id" character varying(36) NOT NULL,
  "price" character varying(78) NOT NULL,
  "address_patterns" text NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_payment_receive_rule_prices_payment_receive_rule" FOREIGN KEY ("payment_receive_rule_id") REFERENCES "public"."payment_receive_rules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
CREATE UNIQUE INDEX "uq_payment_receive_rule_prices_rule_price" ON "public"."payment_receive_rule_prices" ("payment_receive_rule_id", "price");

ALTER TABLE "public"."payment_receive_rule_x402_configs" ADD COLUMN "policy_id" character varying(36) NULL;
-- canonicalize like the API so equal values group
UPDATE "public"."payment_receive_rules"
SET "price" = COALESCE(NULLIF(LTRIM("price", '0'), ''), '0');
UPDATE "public"."payment_receive_rule_x402_configs" AS c
SET "policy_id" = r."policy_id",
    "asset" = LOWER(c."asset"),
    "network" = CASE
      WHEN c."network" LIKE 'eip155:%' AND LTRIM(SUBSTRING(c."network" FROM 8), '0') <> '' THEN 'eip155:' || LTRIM(SUBSTRING(c."network" FROM 8), '0')
      ELSE c."network"
    END
FROM "public"."payment_receive_rules" AS r
WHERE r."id" = c."payment_receive_rule_id";

INSERT INTO "public"."payment_receive_rule_prices" ("id", "payment_receive_rule_id", "price", "address_patterns", "created_at", "updated_at")
WITH "ranked" AS (
  SELECT r."id", r."price", r."created_at", c."address_patterns",
         ROW_NUMBER() OVER (PARTITION BY c."policy_id", c."network", c."asset" ORDER BY r."position", r."created_at", r."id") AS "rank",
         FIRST_VALUE(r."id") OVER (PARTITION BY c."policy_id", c."network", c."asset" ORDER BY r."position", r."created_at", r."id") AS "keeper_id"
  FROM "public"."payment_receive_rules" AS r
  JOIN "public"."payment_receive_rule_x402_configs" AS c ON c."payment_receive_rule_id" = r."id"
),
"patterns" AS (
  SELECT "keeper_id", "price", p."value" AS "pattern", MIN("rank" * 100000 + p."ordinality") AS "sort", MIN("created_at") AS "created_at"
  FROM "ranked", jsonb_array_elements_text("ranked"."address_patterns"::jsonb) WITH ORDINALITY AS p("value", "ordinality")
  GROUP BY "keeper_id", "price", p."value"
)
SELECT gen_random_uuid()::text, "keeper_id", "price", jsonb_agg("pattern" ORDER BY "sort")::text, MIN("created_at"), NOW()
FROM "patterns"
GROUP BY "keeper_id", "price";

DELETE FROM "public"."payment_receive_rules"
WHERE "id" IN (
  SELECT r."id"
  FROM "public"."payment_receive_rules" AS r
  JOIN "public"."payment_receive_rule_x402_configs" AS c ON c."payment_receive_rule_id" = r."id"
  WHERE r."id" <> (
    SELECT r2."id"
    FROM "public"."payment_receive_rules" AS r2
    JOIN "public"."payment_receive_rule_x402_configs" AS c2 ON c2."payment_receive_rule_id" = r2."id"
    WHERE c2."policy_id" = c."policy_id" AND c2."network" = c."network" AND c2."asset" = c."asset"
    ORDER BY r2."position", r2."created_at", r2."id"
    LIMIT 1
  )
);

ALTER TABLE "public"."payment_receive_rule_x402_configs" DROP COLUMN "address_patterns", ALTER COLUMN "policy_id" SET NOT NULL;
CREATE UNIQUE INDEX "uq_payment_receive_rule_x402_configs_policy_asset" ON "public"."payment_receive_rule_x402_configs" ("policy_id", "network", "asset");
ALTER TABLE "public"."payment_receive_rules" DROP COLUMN "position", DROP COLUMN "price";

-- +goose Down
-- Lossy: each rule keeps only its cheapest price.
ALTER TABLE "public"."payment_receive_rules" ADD COLUMN "price" character varying(78) NULL, ADD COLUMN "position" bigint NULL;
ALTER TABLE "public"."payment_receive_rule_x402_configs" ADD COLUMN "address_patterns" text NULL;
UPDATE "public"."payment_receive_rules" AS r
SET "price" = COALESCE((
      SELECT p."price" FROM "public"."payment_receive_rule_prices" AS p
      WHERE p."payment_receive_rule_id" = r."id"
      ORDER BY LENGTH(p."price"), p."price" LIMIT 1
    ), '0'),
    "position" = o."position"
FROM (
  SELECT "id", ROW_NUMBER() OVER (PARTITION BY "policy_id" ORDER BY "created_at", "id") - 1 AS "position"
  FROM "public"."payment_receive_rules"
) AS o
WHERE o."id" = r."id";
UPDATE "public"."payment_receive_rule_x402_configs" AS c
SET "address_patterns" = COALESCE((
  SELECT p."address_patterns" FROM "public"."payment_receive_rule_prices" AS p
  WHERE p."payment_receive_rule_id" = c."payment_receive_rule_id"
  ORDER BY LENGTH(p."price"), p."price" LIMIT 1
), '[]');
ALTER TABLE "public"."payment_receive_rules" ALTER COLUMN "price" SET NOT NULL, ALTER COLUMN "position" SET NOT NULL;
DROP INDEX "public"."uq_payment_receive_rule_x402_configs_policy_asset";
ALTER TABLE "public"."payment_receive_rule_x402_configs" DROP COLUMN "policy_id", ALTER COLUMN "address_patterns" SET NOT NULL;
DROP TABLE "public"."payment_receive_rule_prices";
