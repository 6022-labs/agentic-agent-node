-- +goose Up
-- create "agents" table
CREATE TABLE "public"."agents" (
  "id" character varying(36) NOT NULL,
  "chain_id" character varying(64) NOT NULL,
  "collection_token_id" bigint NULL,
  "collection_address" character varying(42) NULL,
  "proposal_id" bigint NULL,
  "name" character varying(255) NOT NULL,
  "clone_of" bigint NULL,
  "ens_domain" character varying(255) NOT NULL,
  "role" character varying(32) NOT NULL,
  "creator" character varying(42) NOT NULL,
  "wallet_address" character varying(42) NOT NULL,
  "owner" character varying(42) NULL,
  "images" text NULL,
  "attributes" text NULL,
  "updated_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- create "agent_registration_states" table
CREATE TABLE "public"."agent_registration_states" (
  "id" character varying(36) NOT NULL,
  "agent_id" character varying(36) NOT NULL,
  "mint_status" character varying(32) NOT NULL,
  "mint_chain_id" bigint NULL,
  "mint_tx_hash" character varying(66) NULL,
  "minted_at" timestamptz NULL,
  "ens_registry_status" character varying(32) NOT NULL,
  "ens_chain_id" bigint NULL,
  "ens_tx_hash" character varying(66) NULL,
  "ens_published_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_agent_registration_states_agent" FOREIGN KEY ("agent_id") REFERENCES "public"."agents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_agent_registration_states_agent_id" to table: "agent_registration_states"
CREATE UNIQUE INDEX "idx_agent_registration_states_agent_id" ON "public"."agent_registration_states" ("agent_id");
-- create "conversations" table
CREATE TABLE "public"."conversations" (
  "id" character varying(128) NOT NULL,
  "agent_id" character varying(36) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_conversations_agent" FOREIGN KEY ("agent_id") REFERENCES "public"."agents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_conversations_agent_id" to table: "conversations"
CREATE INDEX "idx_conversations_agent_id" ON "public"."conversations" ("agent_id");
-- create "conversation_messages" table
CREATE TABLE "public"."conversation_messages" (
  "id" character varying(36) NOT NULL,
  "conversation_id" character varying(128) NOT NULL,
  "role" character varying(32) NOT NULL,
  "content" text NULL,
  "tool_calls" text NULL,
  "tool_call_id" character varying(128) NULL,
  "hash" character varying(128) NOT NULL,
  "timestamp" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_conversation_messages_conversation" FOREIGN KEY ("conversation_id") REFERENCES "public"."conversations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_conversation_messages_conversation_id" to table: "conversation_messages"
CREATE INDEX "idx_conversation_messages_conversation_id" ON "public"."conversation_messages" ("conversation_id");
-- create index "uq_conversation_messages_hash" to table: "conversation_messages"
CREATE UNIQUE INDEX "uq_conversation_messages_hash" ON "public"."conversation_messages" ("hash");
-- create "llm_connections" table
CREATE TABLE "public"."llm_connections" (
  "id" character varying(36) NOT NULL,
  "agent_id" character varying(36) NOT NULL,
  "platform" character varying(64) NOT NULL,
  "model" character varying(255) NOT NULL,
  "base_url" text NULL,
  "setup_status" character varying(32) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_llm_connections_agent" FOREIGN KEY ("agent_id") REFERENCES "public"."agents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_llm_connections_agent_id" to table: "llm_connections"
CREATE UNIQUE INDEX "idx_llm_connections_agent_id" ON "public"."llm_connections" ("agent_id");
-- create "mcp_connections" table
CREATE TABLE "public"."mcp_connections" (
  "id" character varying(36) NOT NULL,
  "agent_id" character varying(36) NOT NULL,
  "name" character varying(128) NOT NULL,
  "url" text NOT NULL,
  "header_auth_keys" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_mcp_connections_agent" FOREIGN KEY ("agent_id") REFERENCES "public"."agents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "idx_mcp_connections_agent_id" to table: "mcp_connections"
CREATE INDEX "idx_mcp_connections_agent_id" ON "public"."mcp_connections" ("agent_id");
-- create "payment_policies" table
CREATE TABLE "public"."payment_policies" (
  "id" character varying(36) NOT NULL,
  "agent_id" character varying(36) NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_payment_policies_agent" FOREIGN KEY ("agent_id") REFERENCES "public"."agents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
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
  "asset_transfer_method" character varying(16) NOT NULL,
  "eip712_name" character varying(64) NULL,
  "eip712_version" character varying(16) NULL,
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
-- reverse: create index "idx_mcp_connections_agent_id" to table: "mcp_connections"
DROP INDEX "public"."idx_mcp_connections_agent_id";
-- reverse: create "mcp_connections" table
DROP TABLE "public"."mcp_connections";
-- reverse: create index "idx_llm_connections_agent_id" to table: "llm_connections"
DROP INDEX "public"."idx_llm_connections_agent_id";
-- reverse: create "llm_connections" table
DROP TABLE "public"."llm_connections";
-- reverse: create index "uq_conversation_messages_hash" to table: "conversation_messages"
DROP INDEX "public"."uq_conversation_messages_hash";
-- reverse: create index "idx_conversation_messages_conversation_id" to table: "conversation_messages"
DROP INDEX "public"."idx_conversation_messages_conversation_id";
-- reverse: create "conversation_messages" table
DROP TABLE "public"."conversation_messages";
-- reverse: create index "idx_conversations_agent_id" to table: "conversations"
DROP INDEX "public"."idx_conversations_agent_id";
-- reverse: create "conversations" table
DROP TABLE "public"."conversations";
-- reverse: create index "idx_agent_registration_states_agent_id" to table: "agent_registration_states"
DROP INDEX "public"."idx_agent_registration_states_agent_id";
-- reverse: create "agent_registration_states" table
DROP TABLE "public"."agent_registration_states";
-- reverse: create "agents" table
DROP TABLE "public"."agents";
