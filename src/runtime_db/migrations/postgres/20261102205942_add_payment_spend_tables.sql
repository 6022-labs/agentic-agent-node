-- +goose Up
-- create "payment_spend_policies" table
CREATE TABLE "public"."payment_spend_policies" (
  "id" character varying(36) NOT NULL,
  "agent_id" character varying(36) NOT NULL,
  "pay_to_patterns" text NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_payment_spend_policies_agent" FOREIGN KEY ("agent_id") REFERENCES "public"."agents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_payment_spend_policies_agent_id" to table: "payment_spend_policies"
CREATE UNIQUE INDEX "idx_payment_spend_policies_agent_id" ON "public"."payment_spend_policies" ("agent_id");
-- create "payment_spend_rules" table
CREATE TABLE "public"."payment_spend_rules" (
  "id" character varying(36) NOT NULL,
  "policy_id" character varying(36) NOT NULL,
  "name" character varying(64) NOT NULL,
  "method" character varying(32) NOT NULL,
  "max_amount_per_payment" character varying(78) NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_payment_spend_rules_policy" FOREIGN KEY ("policy_id") REFERENCES "public"."payment_spend_policies" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_payment_spend_rules_policy_id" to table: "payment_spend_rules"
CREATE INDEX "idx_payment_spend_rules_policy_id" ON "public"."payment_spend_rules" ("policy_id");
-- create "payment_spend_rule_windows" table
CREATE TABLE "public"."payment_spend_rule_windows" (
  "id" character varying(36) NOT NULL,
  "payment_spend_rule_id" character varying(36) NOT NULL,
  "window_seconds" bigint NOT NULL,
  "max_amount" character varying(78) NOT NULL,
  "spent" character varying(78) NOT NULL DEFAULT '0',
  "started_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_payment_spend_rule_windows_payment_spend_rule" FOREIGN KEY ("payment_spend_rule_id") REFERENCES "public"."payment_spend_rules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "uq_payment_spend_rule_windows_rule_seconds" to table: "payment_spend_rule_windows"
CREATE UNIQUE INDEX "uq_payment_spend_rule_windows_rule_seconds" ON "public"."payment_spend_rule_windows" ("payment_spend_rule_id", "window_seconds");
-- create "payment_spend_rule_x402_configs" table
CREATE TABLE "public"."payment_spend_rule_x402_configs" (
  "payment_spend_rule_id" character varying(36) NOT NULL,
  "policy_id" character varying(36) NOT NULL,
  "network" character varying(64) NOT NULL,
  "asset" character varying(42) NOT NULL,
  "asset_transfer_method" character varying(16) NULL,
  PRIMARY KEY ("payment_spend_rule_id"),
  CONSTRAINT "fk_payment_spend_rule_x402_configs_payment_spend_rule" FOREIGN KEY ("payment_spend_rule_id") REFERENCES "public"."payment_spend_rules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "uq_payment_spend_rule_x402_configs_policy_asset" to table: "payment_spend_rule_x402_configs"
CREATE UNIQUE INDEX "uq_payment_spend_rule_x402_configs_policy_asset" ON "public"."payment_spend_rule_x402_configs" ("policy_id", "network", "asset");

-- +goose Down
-- reverse: create index "uq_payment_spend_rule_x402_configs_policy_asset" to table: "payment_spend_rule_x402_configs"
DROP INDEX "public"."uq_payment_spend_rule_x402_configs_policy_asset";
-- reverse: create "payment_spend_rule_x402_configs" table
DROP TABLE "public"."payment_spend_rule_x402_configs";
-- reverse: create index "uq_payment_spend_rule_windows_rule_seconds" to table: "payment_spend_rule_windows"
DROP INDEX "public"."uq_payment_spend_rule_windows_rule_seconds";
-- reverse: create "payment_spend_rule_windows" table
DROP TABLE "public"."payment_spend_rule_windows";
-- reverse: create index "idx_payment_spend_rules_policy_id" to table: "payment_spend_rules"
DROP INDEX "public"."idx_payment_spend_rules_policy_id";
-- reverse: create "payment_spend_rules" table
DROP TABLE "public"."payment_spend_rules";
-- reverse: create index "idx_payment_spend_policies_agent_id" to table: "payment_spend_policies"
DROP INDEX "public"."idx_payment_spend_policies_agent_id";
-- reverse: create "payment_spend_policies" table
DROP TABLE "public"."payment_spend_policies";
