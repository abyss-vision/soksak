-- +goose Up
CREATE TABLE "agent_task_sessions" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"agent_uuid" uuid NOT NULL,
	"adapter_type" text NOT NULL,
	"task_key" text NOT NULL,
	"session_params_json" jsonb,
	"session_display_id" text,
	"last_run_uuid" uuid,
	"last_error" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "agent_task_sessions" ADD CONSTRAINT "agent_task_sessions_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "agent_task_sessions" ADD CONSTRAINT "agent_task_sessions_agent_uuid_fk" FOREIGN KEY ("agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "agent_task_sessions" ADD CONSTRAINT "agent_task_sessions_last_run_uuid_fk" FOREIGN KEY ("last_run_uuid") REFERENCES "public"."heartbeat_runs"("uuid") ON DELETE no action ON UPDATE no action;
CREATE UNIQUE INDEX "agent_task_sessions_company_agent_adapter_task_uniq" ON "agent_task_sessions" USING btree ("company_uuid","agent_uuid","adapter_type","task_key");
CREATE INDEX "agent_task_sessions_company_agent_updated_idx" ON "agent_task_sessions" USING btree ("company_uuid","agent_uuid","updated_at");
CREATE INDEX "agent_task_sessions_company_task_updated_idx" ON "agent_task_sessions" USING btree ("company_uuid","task_key","updated_at");

-- +goose Down
DROP INDEX IF EXISTS "agent_task_sessions_company_task_updated_idx";
DROP INDEX IF EXISTS "agent_task_sessions_company_agent_updated_idx";
DROP INDEX IF EXISTS "agent_task_sessions_company_agent_adapter_task_uniq";
DROP TABLE IF EXISTS "agent_task_sessions";
