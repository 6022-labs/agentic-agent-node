-- +goose Up
-- create "self_tools" table
CREATE TABLE "public"."self_tools" (
  "id" character varying(36) NOT NULL,
  "agent_id" character varying(36) NOT NULL,
  "name" character varying(128) NOT NULL,
  "enabled" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_self_tools_agent" FOREIGN KEY ("agent_id") REFERENCES "public"."agents" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "uq_self_tools_agent_name" to table: "self_tools"
CREATE UNIQUE INDEX "uq_self_tools_agent_name" ON "public"."self_tools" ("agent_id", "name");

-- +goose Down
-- reverse: create index "uq_self_tools_agent_name" to table: "self_tools"
DROP INDEX "public"."uq_self_tools_agent_name";
-- reverse: create "self_tools" table
DROP TABLE "public"."self_tools";
