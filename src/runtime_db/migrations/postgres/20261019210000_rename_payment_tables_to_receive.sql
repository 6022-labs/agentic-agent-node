-- +goose Up
-- Renames only: the payment tables gain a `receive` half so a `spend` half can
-- sit beside them. Written by hand because atlas diffs a rename as DROP+CREATE,
-- which would destroy every deployed agent's pricing rules.
ALTER TABLE "public"."payment_policies" RENAME TO "payment_receive_policies";
ALTER INDEX "public"."payment_policies_pkey" RENAME TO "payment_receive_policies_pkey";
ALTER INDEX "public"."idx_payment_policies_agent_id" RENAME TO "idx_payment_receive_policies_agent_id";
ALTER TABLE "public"."payment_receive_policies" RENAME CONSTRAINT "fk_payment_policies_agent" TO "fk_payment_receive_policies_agent";

ALTER TABLE "public"."payment_rules" RENAME TO "payment_receive_rules";
ALTER INDEX "public"."payment_rules_pkey" RENAME TO "payment_receive_rules_pkey";
ALTER INDEX "public"."idx_payment_rules_policy_id" RENAME TO "idx_payment_receive_rules_policy_id";
ALTER TABLE "public"."payment_receive_rules" RENAME CONSTRAINT "fk_payment_rules_policy" TO "fk_payment_receive_rules_policy";

ALTER TABLE "public"."payment_rule_x402_configs" RENAME TO "payment_receive_rule_x402_configs";
ALTER TABLE "public"."payment_receive_rule_x402_configs" RENAME COLUMN "payment_rule_id" TO "payment_receive_rule_id";
ALTER INDEX "public"."payment_rule_x402_configs_pkey" RENAME TO "payment_receive_rule_x402_configs_pkey";
ALTER TABLE "public"."payment_receive_rule_x402_configs" RENAME CONSTRAINT "fk_payment_rule_x402_configs_payment_rule" TO "fk_payment_receive_rule_x402_configs_payment_receive_rule";

-- the per-agent enabled flag is keyed by tool name
UPDATE "public"."self_tools" SET name = 'create_payment_receive_rule' WHERE name = 'create_payment_rule';
UPDATE "public"."self_tools" SET name = 'update_payment_receive_rule' WHERE name = 'update_payment_rule';
UPDATE "public"."self_tools" SET name = 'delete_payment_receive_rule' WHERE name = 'delete_payment_rule';
UPDATE "public"."self_tools" SET name = 'list_payment_receive_rules' WHERE name = 'list_payment_rules';
UPDATE "public"."self_tools" SET name = 'get_payment_receive_policy' WHERE name = 'get_payment_policy';
UPDATE "public"."self_tools" SET name = 'delete_payment_receive_policy' WHERE name = 'delete_payment_policy';

-- +goose Down
ALTER TABLE "public"."payment_receive_rule_x402_configs" RENAME CONSTRAINT "fk_payment_receive_rule_x402_configs_payment_receive_rule" TO "fk_payment_rule_x402_configs_payment_rule";
ALTER INDEX "public"."payment_receive_rule_x402_configs_pkey" RENAME TO "payment_rule_x402_configs_pkey";
ALTER TABLE "public"."payment_receive_rule_x402_configs" RENAME COLUMN "payment_receive_rule_id" TO "payment_rule_id";
ALTER TABLE "public"."payment_receive_rule_x402_configs" RENAME TO "payment_rule_x402_configs";

ALTER TABLE "public"."payment_receive_rules" RENAME CONSTRAINT "fk_payment_receive_rules_policy" TO "fk_payment_rules_policy";
ALTER INDEX "public"."idx_payment_receive_rules_policy_id" RENAME TO "idx_payment_rules_policy_id";
ALTER INDEX "public"."payment_receive_rules_pkey" RENAME TO "payment_rules_pkey";
ALTER TABLE "public"."payment_receive_rules" RENAME TO "payment_rules";

ALTER TABLE "public"."payment_receive_policies" RENAME CONSTRAINT "fk_payment_receive_policies_agent" TO "fk_payment_policies_agent";
ALTER INDEX "public"."idx_payment_receive_policies_agent_id" RENAME TO "idx_payment_policies_agent_id";
ALTER INDEX "public"."payment_receive_policies_pkey" RENAME TO "payment_policies_pkey";
ALTER TABLE "public"."payment_receive_policies" RENAME TO "payment_policies";

UPDATE "public"."self_tools" SET name = 'create_payment_rule' WHERE name = 'create_payment_receive_rule';
UPDATE "public"."self_tools" SET name = 'update_payment_rule' WHERE name = 'update_payment_receive_rule';
UPDATE "public"."self_tools" SET name = 'delete_payment_rule' WHERE name = 'delete_payment_receive_rule';
UPDATE "public"."self_tools" SET name = 'list_payment_rules' WHERE name = 'list_payment_receive_rules';
UPDATE "public"."self_tools" SET name = 'get_payment_policy' WHERE name = 'get_payment_receive_policy';
UPDATE "public"."self_tools" SET name = 'delete_payment_policy' WHERE name = 'delete_payment_receive_policy';
