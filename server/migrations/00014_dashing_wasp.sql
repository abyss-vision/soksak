-- +goose Up
ALTER TABLE "issues" ADD COLUMN "execution_run_uuid" uuid;
ALTER TABLE "issues" ADD COLUMN "execution_agent_name_key" text;
ALTER TABLE "issues" ADD COLUMN "execution_locked_at" timestamp with time zone;
ALTER TABLE "issues" ADD CONSTRAINT "issues_execution_run_uuid_fk" FOREIGN KEY ("execution_run_uuid") REFERENCES "public"."heartbeat_runs"("uuid") ON DELETE set null ON UPDATE no action;

-- +goose Down
ALTER TABLE "issues" DROP CONSTRAINT IF EXISTS "issues_execution_run_uuid_fk";
ALTER TABLE "issues" DROP COLUMN IF EXISTS "execution_locked_at";
ALTER TABLE "issues" DROP COLUMN IF EXISTS "execution_agent_name_key";
ALTER TABLE "issues" DROP COLUMN IF EXISTS "execution_run_uuid";
