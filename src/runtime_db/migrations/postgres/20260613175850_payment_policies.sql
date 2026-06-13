-- +goose Up
-- create "payment_policies" table
CREATE TABLE "public"."payment_policies" (
  "id" character varying(36) NOT NULL,
  "agent_id" character varying(36) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_payment_policies_agent_id" to table: "payment_policies"
CREATE UNIQUE INDEX "idx_payment_policies_agent_id" ON "public"."payment_policies" ("agent_id");
-- create "payment_rules" table
CREATE TABLE "public"."payment_rules" (
  "id" character varying(36) NOT NULL,
  "policy_id" character varying(36) NOT NULL,
  "position" bigint NOT NULL,
  "name" character varying(64) NOT NULL,
  "price" character varying(78) NOT NULL,
  "method" character varying(32) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_payment_rules_policy" FOREIGN KEY ("policy_id") REFERENCES "public"."payment_policies" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_payment_rules_policy_id" to table: "payment_rules"
CREATE INDEX "idx_payment_rules_policy_id" ON "public"."payment_rules" ("policy_id");
-- create "payment_rule_x402_configs" table
CREATE TABLE "public"."payment_rule_x402_configs" (
  "payment_rule_id" character varying(36) NOT NULL,
  "network" character varying(64) NOT NULL,
  "asset" character varying(42) NOT NULL,
  "pay_to" character varying(42) NOT NULL,
  "eip712_name" character varying(64) NOT NULL,
  "eip712_version" character varying(16) NOT NULL,
  "address_patterns" text NOT NULL,
  PRIMARY KEY ("payment_rule_id"),
  CONSTRAINT "fk_payment_rule_x402_configs_payment_rule" FOREIGN KEY ("payment_rule_id") REFERENCES "public"."payment_rules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "payment_rule_x402_configs" table
DROP TABLE "public"."payment_rule_x402_configs";
-- reverse: create index "idx_payment_rules_policy_id" to table: "payment_rules"
DROP INDEX "public"."idx_payment_rules_policy_id";
-- reverse: create "payment_rules" table
DROP TABLE "public"."payment_rules";
-- reverse: create index "idx_payment_policies_agent_id" to table: "payment_policies"
DROP INDEX "public"."idx_payment_policies_agent_id";
-- reverse: create "payment_policies" table
DROP TABLE "public"."payment_policies";
