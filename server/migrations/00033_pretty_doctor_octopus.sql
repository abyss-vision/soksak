-- +goose Up
CREATE TABLE "budget_incidents" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"policy_uuid" uuid NOT NULL,
	"scope_type" text NOT NULL,
	"scope_uuid" uuid NOT NULL,
	"metric" text NOT NULL,
	"window_kind" text NOT NULL,
	"window_start" timestamp with time zone NOT NULL,
	"window_end" timestamp with time zone NOT NULL,
	"threshold_type" text NOT NULL,
	"amount_limit" integer NOT NULL,
	"amount_observed" integer NOT NULL,
	"status" text DEFAULT 'open' NOT NULL,
	"approval_uuid" uuid,
	"resolved_at" timestamp with time zone,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "budget_policies" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"scope_type" text NOT NULL,
	"scope_uuid" uuid NOT NULL,
	"metric" text DEFAULT 'billed_cents' NOT NULL,
	"window_kind" text NOT NULL,
	"amount" integer DEFAULT 0 NOT NULL,
	"warn_percent" integer DEFAULT 80 NOT NULL,
	"hard_stop_enabled" boolean DEFAULT true NOT NULL,
	"notify_enabled" boolean DEFAULT true NOT NULL,
	"is_active" boolean DEFAULT true NOT NULL,
	"created_by_user_id" text,
	"updated_by_user_id" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "agents" ADD COLUMN "pause_reason" text;
ALTER TABLE "agents" ADD COLUMN "paused_at" timestamp with time zone;
ALTER TABLE "projects" ADD COLUMN "pause_reason" text;
ALTER TABLE "projects" ADD COLUMN "paused_at" timestamp with time zone;
INSERT INTO "budget_policies" (
	"company_uuid",
	"scope_type",
	"scope_uuid",
	"metric",
	"window_kind",
	"amount",
	"warn_percent",
	"hard_stop_enabled",
	"notify_enabled",
	"is_active"
)
SELECT
	"uuid",
	'company',
	"uuid",
	'billed_cents',
	'calendar_month_utc',
	"budget_monthly_cents",
	80,
	true,
	true,
	true
FROM "companies"
WHERE "budget_monthly_cents" > 0;
INSERT INTO "budget_policies" (
	"company_uuid",
	"scope_type",
	"scope_uuid",
	"metric",
	"window_kind",
	"amount",
	"warn_percent",
	"hard_stop_enabled",
	"notify_enabled",
	"is_active"
)
SELECT
	"company_uuid",
	'agent',
	"uuid",
	'billed_cents',
	'calendar_month_utc',
	"budget_monthly_cents",
	80,
	true,
	true,
	true
FROM "agents"
WHERE "budget_monthly_cents" > 0;
ALTER TABLE "budget_incidents" ADD CONSTRAINT "budget_incidents_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "budget_incidents" ADD CONSTRAINT "budget_incidents_policy_uuid_fk" FOREIGN KEY ("policy_uuid") REFERENCES "public"."budget_policies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "budget_incidents" ADD CONSTRAINT "budget_incidents_approval_uuid_fk" FOREIGN KEY ("approval_uuid") REFERENCES "public"."approvals"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "budget_policies" ADD CONSTRAINT "budget_policies_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
CREATE INDEX "budget_incidents_company_status_idx" ON "budget_incidents" USING btree ("company_uuid","status");
CREATE INDEX "budget_incidents_company_scope_idx" ON "budget_incidents" USING btree ("company_uuid","scope_type","scope_uuid","status");
CREATE UNIQUE INDEX "budget_incidents_policy_window_threshold_idx" ON "budget_incidents" USING btree ("policy_uuid","window_start","threshold_type");
CREATE INDEX "budget_policies_company_scope_active_idx" ON "budget_policies" USING btree ("company_uuid","scope_type","scope_uuid","is_active");
CREATE INDEX "budget_policies_company_window_idx" ON "budget_policies" USING btree ("company_uuid","window_kind","metric");
CREATE UNIQUE INDEX "budget_policies_company_scope_metric_unique_idx" ON "budget_policies" USING btree ("company_uuid","scope_type","scope_uuid","metric","window_kind");

-- +goose Down
DROP TABLE IF EXISTS "budget_incidents";
DROP TABLE IF EXISTS "budget_policies";
ALTER TABLE "agents" DROP COLUMN IF EXISTS "paused_at";
ALTER TABLE "agents" DROP COLUMN IF EXISTS "pause_reason";
ALTER TABLE "projects" DROP COLUMN IF EXISTS "paused_at";
ALTER TABLE "projects" DROP COLUMN IF EXISTS "pause_reason";
