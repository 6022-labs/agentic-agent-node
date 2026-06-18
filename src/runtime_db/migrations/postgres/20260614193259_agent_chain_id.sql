-- +goose Up
-- modify "agents" table
ALTER TABLE "public"."agents" ADD COLUMN "chain_id" character varying(64) NOT NULL;

-- +goose Down
-- reverse: modify "agents" table
ALTER TABLE "public"."agents" DROP COLUMN "chain_id";
