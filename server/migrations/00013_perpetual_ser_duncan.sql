-- +goose Up
ALTER TABLE "issues" ADD COLUMN "checkout_run_uuid" uuid;
ALTER TABLE "issues" ADD CONSTRAINT "issues_checkout_run_uuid_fk" FOREIGN KEY ("checkout_run_uuid") REFERENCES "public"."heartbeat_runs"("uuid") ON DELETE set null ON UPDATE no action;

-- +goose Down
ALTER TABLE "issues" DROP CONSTRAINT IF EXISTS "issues_checkout_run_uuid_fk";
ALTER TABLE "issues" DROP COLUMN IF EXISTS "checkout_run_uuid";
