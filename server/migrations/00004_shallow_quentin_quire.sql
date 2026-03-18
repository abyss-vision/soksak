-- +goose Up
ALTER TABLE "activity_log" ADD COLUMN "run_uuid" uuid;
ALTER TABLE "activity_log" ADD CONSTRAINT "activity_log_run_uuid_fk" FOREIGN KEY ("run_uuid") REFERENCES "public"."heartbeat_runs"("uuid") ON DELETE no action ON UPDATE no action;
CREATE INDEX "activity_log_run_uuid_idx" ON "activity_log" USING btree ("run_uuid");
CREATE INDEX "activity_log_entity_type_id_idx" ON "activity_log" USING btree ("entity_type","entity_id");
ALTER TABLE "agents" DROP COLUMN "context_mode";

-- +goose Down
ALTER TABLE "agents" ADD COLUMN "context_mode" text DEFAULT 'thin' NOT NULL;
DROP INDEX IF EXISTS "activity_log_entity_type_id_idx";
DROP INDEX IF EXISTS "activity_log_run_uuid_idx";
ALTER TABLE "activity_log" DROP CONSTRAINT IF EXISTS "activity_log_run_uuid_fk";
ALTER TABLE "activity_log" DROP COLUMN IF EXISTS "run_uuid";
