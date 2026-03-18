-- +goose Up
CREATE TABLE "agent_runtime_state" (
	"agent_uuid" uuid PRIMARY KEY NOT NULL,
	"company_uuid" uuid NOT NULL,
	"adapter_type" text NOT NULL,
	"session_id" text,
	"state_json" jsonb DEFAULT '{}'::jsonb NOT NULL,
	"last_run_uuid" uuid,
	"last_run_status" text,
	"total_input_tokens" bigint DEFAULT 0 NOT NULL,
	"total_output_tokens" bigint DEFAULT 0 NOT NULL,
	"total_cached_input_tokens" bigint DEFAULT 0 NOT NULL,
	"total_cost_cents" bigint DEFAULT 0 NOT NULL,
	"last_error" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "agent_wakeup_requests" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"agent_uuid" uuid NOT NULL,
	"source" text NOT NULL,
	"trigger_detail" text,
	"reason" text,
	"payload" jsonb,
	"status" text DEFAULT 'queued' NOT NULL,
	"coalesced_count" integer DEFAULT 0 NOT NULL,
	"requested_by_actor_type" text,
	"requested_by_actor_id" text,
	"idempotency_key" text,
	"run_uuid" uuid,
	"requested_at" timestamp with time zone DEFAULT now() NOT NULL,
	"claimed_at" timestamp with time zone,
	"finished_at" timestamp with time zone,
	"error" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "heartbeat_run_events" (
	"id" bigserial PRIMARY KEY NOT NULL,
	"company_uuid" uuid NOT NULL,
	"run_uuid" uuid NOT NULL,
	"agent_uuid" uuid NOT NULL,
	"seq" integer NOT NULL,
	"event_type" text NOT NULL,
	"stream" text,
	"level" text,
	"color" text,
	"message" text,
	"payload" jsonb,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "heartbeat_runs" ALTER COLUMN "invocation_source" SET DEFAULT 'on_demand';
ALTER TABLE "agents" ADD COLUMN "runtime_config" jsonb DEFAULT '{}'::jsonb NOT NULL;
ALTER TABLE "heartbeat_runs" ADD COLUMN "trigger_detail" text;
ALTER TABLE "heartbeat_runs" ADD COLUMN "wakeup_request_uuid" uuid;
ALTER TABLE "heartbeat_runs" ADD COLUMN "exit_code" integer;
ALTER TABLE "heartbeat_runs" ADD COLUMN "signal" text;
ALTER TABLE "heartbeat_runs" ADD COLUMN "usage_json" jsonb;
ALTER TABLE "heartbeat_runs" ADD COLUMN "result_json" jsonb;
ALTER TABLE "heartbeat_runs" ADD COLUMN "session_id_before" text;
ALTER TABLE "heartbeat_runs" ADD COLUMN "session_id_after" text;
ALTER TABLE "heartbeat_runs" ADD COLUMN "log_store" text;
ALTER TABLE "heartbeat_runs" ADD COLUMN "log_ref" text;
ALTER TABLE "heartbeat_runs" ADD COLUMN "log_bytes" bigint;
ALTER TABLE "heartbeat_runs" ADD COLUMN "log_sha256" text;
ALTER TABLE "heartbeat_runs" ADD COLUMN "log_compressed" boolean DEFAULT false NOT NULL;
ALTER TABLE "heartbeat_runs" ADD COLUMN "stdout_excerpt" text;
ALTER TABLE "heartbeat_runs" ADD COLUMN "stderr_excerpt" text;
ALTER TABLE "heartbeat_runs" ADD COLUMN "error_code" text;
ALTER TABLE "agent_runtime_state" ADD CONSTRAINT "agent_runtime_state_agent_uuid_fk" FOREIGN KEY ("agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "agent_runtime_state" ADD CONSTRAINT "agent_runtime_state_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "agent_wakeup_requests" ADD CONSTRAINT "agent_wakeup_requests_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "agent_wakeup_requests" ADD CONSTRAINT "agent_wakeup_requests_agent_uuid_fk" FOREIGN KEY ("agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "heartbeat_run_events" ADD CONSTRAINT "heartbeat_run_events_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "heartbeat_run_events" ADD CONSTRAINT "heartbeat_run_events_run_uuid_fk" FOREIGN KEY ("run_uuid") REFERENCES "public"."heartbeat_runs"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "heartbeat_run_events" ADD CONSTRAINT "heartbeat_run_events_agent_uuid_fk" FOREIGN KEY ("agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
CREATE INDEX "agent_runtime_state_company_agent_idx" ON "agent_runtime_state" USING btree ("company_uuid","agent_uuid");
CREATE INDEX "agent_runtime_state_company_updated_idx" ON "agent_runtime_state" USING btree ("company_uuid","updated_at");
CREATE INDEX "agent_wakeup_requests_company_agent_status_idx" ON "agent_wakeup_requests" USING btree ("company_uuid","agent_uuid","status");
CREATE INDEX "agent_wakeup_requests_company_requested_idx" ON "agent_wakeup_requests" USING btree ("company_uuid","requested_at");
CREATE INDEX "agent_wakeup_requests_agent_requested_idx" ON "agent_wakeup_requests" USING btree ("agent_uuid","requested_at");
CREATE INDEX "heartbeat_run_events_run_seq_idx" ON "heartbeat_run_events" USING btree ("run_uuid","seq");
CREATE INDEX "heartbeat_run_events_company_run_idx" ON "heartbeat_run_events" USING btree ("company_uuid","run_uuid");
CREATE INDEX "heartbeat_run_events_company_created_idx" ON "heartbeat_run_events" USING btree ("company_uuid","created_at");

-- +goose Down
DROP TABLE IF EXISTS "heartbeat_run_events";
DROP TABLE IF EXISTS "agent_wakeup_requests";
DROP TABLE IF EXISTS "agent_runtime_state";
