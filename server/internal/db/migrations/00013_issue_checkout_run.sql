-- +goose Up
ALTER TABLE "issues" ADD COLUMN "checkout_run_id" uuid;
ALTER TABLE "issues" ADD CONSTRAINT "issues_checkout_run_id_heartbeat_runs_id_fk" FOREIGN KEY ("checkout_run_id") REFERENCES "public"."heartbeat_runs"("id") ON DELETE set null ON UPDATE no action;

-- +goose Down
ALTER TABLE "issues" DROP CONSTRAINT IF EXISTS "issues_checkout_run_id_heartbeat_runs_id_fk";
ALTER TABLE "issues" DROP COLUMN IF EXISTS "checkout_run_id";
