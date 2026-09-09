-- +goose Up
-- create "payment_spend_policies" table
CREATE TABLE "public"."payment_spend_policies" (
  "id" character varying(36) NOT NULL,
  "agent_id" character varying(36) NOT NULL,
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
  "position" bigint NOT NULL,
  "name" character varying(64) NOT NULL,
  "max_amount_per_payment" character varying(78) NOT NULL,
  "method" character varying(32) NOT NULL,
  "window_seconds" bigint NOT NULL DEFAULT 0,
  "max_amount_per_window" character varying(78) NOT NULL DEFAULT '0',
  "window_spent" character varying(78) NOT NULL DEFAULT '0',
  "window_started_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_payment_spend_rules_policy" FOREIGN KEY ("policy_id") REFERENCES "public"."payment_spend_policies" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_payment_spend_rules_policy_id" to table: "payment_spend_rules"
CREATE INDEX "idx_payment_spend_rules_policy_id" ON "public"."payment_spend_rules" ("policy_id");
-- create "payment_spend_rule_x402_configs" table
CREATE TABLE "public"."payment_spend_rule_x402_configs" (
  "payment_spend_rule_id" character varying(36) NOT NULL,
  "network" character varying(64) NOT NULL,
  "asset" character varying(42) NOT NULL,
  "asset_transfer_method" character varying(16) NULL,
  "pay_to_patterns" text NOT NULL,
  PRIMARY KEY ("payment_spend_rule_id"),
  CONSTRAINT "fk_payment_spend_rule_x402_configs_payment_spend_rule" FOREIGN KEY ("payment_spend_rule_id") REFERENCES "public"."payment_spend_rules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "payment_spend_rule_x402_configs" table
DROP TABLE "public"."payment_spend_rule_x402_configs";
-- reverse: create index "idx_payment_spend_rules_policy_id" to table: "payment_spend_rules"
DROP INDEX "public"."idx_payment_spend_rules_policy_id";
-- reverse: create "payment_spend_rules" table
DROP TABLE "public"."payment_spend_rules";
-- reverse: create index "idx_payment_spend_policies_agent_id" to table: "payment_spend_policies"
DROP INDEX "public"."idx_payment_spend_policies_agent_id";
-- reverse: create "payment_spend_policies" table
DROP TABLE "public"."payment_spend_policies";
