-- +goose Up
CREATE TABLE "finance_events" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"agent_uuid" uuid,
	"issue_uuid" uuid,
	"project_uuid" uuid,
	"goal_uuid" uuid,
	"heartbeat_run_uuid" uuid,
	"cost_event_uuid" uuid,
	"billing_code" text,
	"description" text,
	"event_kind" text NOT NULL,
	"direction" text DEFAULT 'debit' NOT NULL,
	"biller" text NOT NULL,
	"provider" text,
	"execution_adapter_type" text,
	"pricing_tier" text,
	"region" text,
	"model" text,
	"quantity" integer,
	"unit" text,
	"amount_cents" integer NOT NULL,
	"currency" text DEFAULT 'USD' NOT NULL,
	"estimated" boolean DEFAULT false NOT NULL,
	"external_invoice_id" text,
	"metadata_json" jsonb,
	"occurred_at" timestamp with time zone NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "cost_events" ADD COLUMN "heartbeat_run_uuid" uuid;
ALTER TABLE "cost_events" ADD COLUMN "biller" text DEFAULT 'unknown' NOT NULL;
ALTER TABLE "cost_events" ADD COLUMN "billing_type" text DEFAULT 'unknown' NOT NULL;
ALTER TABLE "cost_events" ADD COLUMN "cached_input_tokens" integer DEFAULT 0 NOT NULL;
ALTER TABLE "finance_events" ADD CONSTRAINT "finance_events_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "finance_events" ADD CONSTRAINT "finance_events_agent_uuid_fk" FOREIGN KEY ("agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "finance_events" ADD CONSTRAINT "finance_events_issue_uuid_fk" FOREIGN KEY ("issue_uuid") REFERENCES "public"."issues"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "finance_events" ADD CONSTRAINT "finance_events_project_uuid_fk" FOREIGN KEY ("project_uuid") REFERENCES "public"."projects"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "finance_events" ADD CONSTRAINT "finance_events_goal_uuid_fk" FOREIGN KEY ("goal_uuid") REFERENCES "public"."goals"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "finance_events" ADD CONSTRAINT "finance_events_heartbeat_run_uuid_fk" FOREIGN KEY ("heartbeat_run_uuid") REFERENCES "public"."heartbeat_runs"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "finance_events" ADD CONSTRAINT "finance_events_cost_event_uuid_fk" FOREIGN KEY ("cost_event_uuid") REFERENCES "public"."cost_events"("uuid") ON DELETE no action ON UPDATE no action;
CREATE INDEX "finance_events_company_occurred_idx" ON "finance_events" USING btree ("company_uuid","occurred_at");
CREATE INDEX "finance_events_company_biller_occurred_idx" ON "finance_events" USING btree ("company_uuid","biller","occurred_at");
CREATE INDEX "finance_events_company_kind_occurred_idx" ON "finance_events" USING btree ("company_uuid","event_kind","occurred_at");
CREATE INDEX "finance_events_company_direction_occurred_idx" ON "finance_events" USING btree ("company_uuid","direction","occurred_at");
CREATE INDEX "finance_events_company_heartbeat_run_idx" ON "finance_events" USING btree ("company_uuid","heartbeat_run_uuid");
CREATE INDEX "finance_events_company_cost_event_idx" ON "finance_events" USING btree ("company_uuid","cost_event_uuid");
ALTER TABLE "cost_events" ADD CONSTRAINT "cost_events_heartbeat_run_uuid_fk" FOREIGN KEY ("heartbeat_run_uuid") REFERENCES "public"."heartbeat_runs"("uuid") ON DELETE no action ON UPDATE no action;
CREATE INDEX "cost_events_company_provider_occurred_idx" ON "cost_events" USING btree ("company_uuid","provider","occurred_at");
CREATE INDEX "cost_events_company_biller_occurred_idx" ON "cost_events" USING btree ("company_uuid","biller","occurred_at");
CREATE INDEX "cost_events_company_heartbeat_run_idx" ON "cost_events" USING btree ("company_uuid","heartbeat_run_uuid");

-- +goose Down
DROP TABLE IF EXISTS "finance_events";
ALTER TABLE "cost_events" DROP CONSTRAINT IF EXISTS "cost_events_heartbeat_run_uuid_fk";
ALTER TABLE "cost_events" DROP COLUMN IF EXISTS "cached_input_tokens";
ALTER TABLE "cost_events" DROP COLUMN IF EXISTS "billing_type";
ALTER TABLE "cost_events" DROP COLUMN IF EXISTS "biller";
ALTER TABLE "cost_events" DROP COLUMN IF EXISTS "heartbeat_run_uuid";
