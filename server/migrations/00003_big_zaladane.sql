-- +goose Up
ALTER TABLE "heartbeat_runs" ADD CONSTRAINT "heartbeat_runs_wakeup_request_uuid_fk" FOREIGN KEY ("wakeup_request_uuid") REFERENCES "public"."agent_wakeup_requests"("uuid") ON DELETE no action ON UPDATE no action;

-- +goose Down
ALTER TABLE "heartbeat_runs" DROP CONSTRAINT IF EXISTS "heartbeat_runs_wakeup_request_uuid_fk";
