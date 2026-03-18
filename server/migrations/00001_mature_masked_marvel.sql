-- +goose Up
CREATE TABLE "activity_log" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"actor_type" text DEFAULT 'system' NOT NULL,
	"actor_id" text NOT NULL,
	"action" text NOT NULL,
	"entity_type" text NOT NULL,
	"entity_id" text NOT NULL,
	"agent_uuid" uuid,
	"details" jsonb,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "agent_api_keys" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"agent_uuid" uuid NOT NULL,
	"company_uuid" uuid NOT NULL,
	"name" text NOT NULL,
	"key_hash" text NOT NULL,
	"last_used_at" timestamp with time zone,
	"revoked_at" timestamp with time zone,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "agents" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"name" text NOT NULL,
	"role" text DEFAULT 'general' NOT NULL,
	"title" text,
	"status" text DEFAULT 'idle' NOT NULL,
	"reports_to" uuid,
	"capabilities" text,
	"adapter_type" text DEFAULT 'process' NOT NULL,
	"adapter_config" jsonb DEFAULT '{}'::jsonb NOT NULL,
	"context_mode" text DEFAULT 'thin' NOT NULL,
	"budget_monthly_cents" integer DEFAULT 0 NOT NULL,
	"spent_monthly_cents" integer DEFAULT 0 NOT NULL,
	"last_heartbeat_at" timestamp with time zone,
	"metadata" jsonb,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "approvals" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"type" text NOT NULL,
	"requested_by_agent_uuid" uuid,
	"requested_by_user_id" text,
	"status" text DEFAULT 'pending' NOT NULL,
	"payload" jsonb NOT NULL,
	"decision_note" text,
	"decided_by_user_id" text,
	"decided_at" timestamp with time zone,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "companies" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"name" text NOT NULL,
	"description" text,
	"status" text DEFAULT 'active' NOT NULL,
	"budget_monthly_cents" integer DEFAULT 0 NOT NULL,
	"spent_monthly_cents" integer DEFAULT 0 NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "cost_events" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"agent_uuid" uuid NOT NULL,
	"issue_uuid" uuid,
	"project_uuid" uuid,
	"goal_uuid" uuid,
	"billing_code" text,
	"provider" text NOT NULL,
	"model" text NOT NULL,
	"input_tokens" integer DEFAULT 0 NOT NULL,
	"output_tokens" integer DEFAULT 0 NOT NULL,
	"cost_cents" integer NOT NULL,
	"occurred_at" timestamp with time zone NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "goals" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"title" text NOT NULL,
	"description" text,
	"level" text DEFAULT 'task' NOT NULL,
	"status" text DEFAULT 'planned' NOT NULL,
	"parent_uuid" uuid,
	"owner_agent_uuid" uuid,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "heartbeat_runs" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"agent_uuid" uuid NOT NULL,
	"invocation_source" text DEFAULT 'manual' NOT NULL,
	"status" text DEFAULT 'queued' NOT NULL,
	"started_at" timestamp with time zone,
	"finished_at" timestamp with time zone,
	"error" text,
	"external_run_id" text,
	"context_snapshot" jsonb,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "issue_comments" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"issue_uuid" uuid NOT NULL,
	"author_agent_uuid" uuid,
	"author_user_id" text,
	"body" text NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "issues" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"project_uuid" uuid,
	"goal_uuid" uuid,
	"parent_uuid" uuid,
	"title" text NOT NULL,
	"description" text,
	"status" text DEFAULT 'backlog' NOT NULL,
	"priority" text DEFAULT 'medium' NOT NULL,
	"assignee_agent_uuid" uuid,
	"created_by_agent_uuid" uuid,
	"created_by_user_id" text,
	"request_depth" integer DEFAULT 0 NOT NULL,
	"billing_code" text,
	"started_at" timestamp with time zone,
	"completed_at" timestamp with time zone,
	"cancelled_at" timestamp with time zone,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "projects" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"goal_uuid" uuid,
	"name" text NOT NULL,
	"description" text,
	"status" text DEFAULT 'backlog' NOT NULL,
	"lead_agent_uuid" uuid,
	"target_date" date,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "activity_log" ADD CONSTRAINT "activity_log_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "activity_log" ADD CONSTRAINT "activity_log_agent_uuid_fk" FOREIGN KEY ("agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "agent_api_keys" ADD CONSTRAINT "agent_api_keys_agent_uuid_fk" FOREIGN KEY ("agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "agent_api_keys" ADD CONSTRAINT "agent_api_keys_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "agents" ADD CONSTRAINT "agents_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "agents" ADD CONSTRAINT "agents_reports_to_fk" FOREIGN KEY ("reports_to") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "approvals" ADD CONSTRAINT "approvals_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "approvals" ADD CONSTRAINT "approvals_requested_by_agent_uuid_fk" FOREIGN KEY ("requested_by_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "cost_events" ADD CONSTRAINT "cost_events_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "cost_events" ADD CONSTRAINT "cost_events_agent_uuid_fk" FOREIGN KEY ("agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "cost_events" ADD CONSTRAINT "cost_events_issue_uuid_fk" FOREIGN KEY ("issue_uuid") REFERENCES "public"."issues"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "cost_events" ADD CONSTRAINT "cost_events_project_uuid_fk" FOREIGN KEY ("project_uuid") REFERENCES "public"."projects"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "cost_events" ADD CONSTRAINT "cost_events_goal_uuid_fk" FOREIGN KEY ("goal_uuid") REFERENCES "public"."goals"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "goals" ADD CONSTRAINT "goals_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "goals" ADD CONSTRAINT "goals_parent_uuid_fk" FOREIGN KEY ("parent_uuid") REFERENCES "public"."goals"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "goals" ADD CONSTRAINT "goals_owner_agent_uuid_fk" FOREIGN KEY ("owner_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "heartbeat_runs" ADD CONSTRAINT "heartbeat_runs_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "heartbeat_runs" ADD CONSTRAINT "heartbeat_runs_agent_uuid_fk" FOREIGN KEY ("agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issue_comments" ADD CONSTRAINT "issue_comments_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issue_comments" ADD CONSTRAINT "issue_comments_issue_uuid_fk" FOREIGN KEY ("issue_uuid") REFERENCES "public"."issues"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issue_comments" ADD CONSTRAINT "issue_comments_author_agent_uuid_fk" FOREIGN KEY ("author_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issues" ADD CONSTRAINT "issues_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issues" ADD CONSTRAINT "issues_project_uuid_fk" FOREIGN KEY ("project_uuid") REFERENCES "public"."projects"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issues" ADD CONSTRAINT "issues_goal_uuid_fk" FOREIGN KEY ("goal_uuid") REFERENCES "public"."goals"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issues" ADD CONSTRAINT "issues_parent_uuid_fk" FOREIGN KEY ("parent_uuid") REFERENCES "public"."issues"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issues" ADD CONSTRAINT "issues_assignee_agent_uuid_fk" FOREIGN KEY ("assignee_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issues" ADD CONSTRAINT "issues_created_by_agent_uuid_fk" FOREIGN KEY ("created_by_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "projects" ADD CONSTRAINT "projects_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "projects" ADD CONSTRAINT "projects_goal_uuid_fk" FOREIGN KEY ("goal_uuid") REFERENCES "public"."goals"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "projects" ADD CONSTRAINT "projects_lead_agent_uuid_fk" FOREIGN KEY ("lead_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
CREATE INDEX "activity_log_company_created_idx" ON "activity_log" USING btree ("company_uuid","created_at");
CREATE INDEX "agent_api_keys_key_hash_idx" ON "agent_api_keys" USING btree ("key_hash");
CREATE INDEX "agent_api_keys_company_agent_idx" ON "agent_api_keys" USING btree ("company_uuid","agent_uuid");
CREATE INDEX "agents_company_status_idx" ON "agents" USING btree ("company_uuid","status");
CREATE INDEX "agents_company_reports_to_idx" ON "agents" USING btree ("company_uuid","reports_to");
CREATE INDEX "approvals_company_status_type_idx" ON "approvals" USING btree ("company_uuid","status","type");
CREATE INDEX "cost_events_company_occurred_idx" ON "cost_events" USING btree ("company_uuid","occurred_at");
CREATE INDEX "cost_events_company_agent_occurred_idx" ON "cost_events" USING btree ("company_uuid","agent_uuid","occurred_at");
CREATE INDEX "goals_company_idx" ON "goals" USING btree ("company_uuid");
CREATE INDEX "heartbeat_runs_company_agent_started_idx" ON "heartbeat_runs" USING btree ("company_uuid","agent_uuid","started_at");
CREATE INDEX "issue_comments_issue_idx" ON "issue_comments" USING btree ("issue_uuid");
CREATE INDEX "issue_comments_company_idx" ON "issue_comments" USING btree ("company_uuid");
CREATE INDEX "issues_company_status_idx" ON "issues" USING btree ("company_uuid","status");
CREATE INDEX "issues_company_assignee_status_idx" ON "issues" USING btree ("company_uuid","assignee_agent_uuid","status");
CREATE INDEX "issues_company_parent_idx" ON "issues" USING btree ("company_uuid","parent_uuid");
CREATE INDEX "issues_company_project_idx" ON "issues" USING btree ("company_uuid","project_uuid");
CREATE INDEX "projects_company_idx" ON "projects" USING btree ("company_uuid");

-- +goose Down
DROP TABLE IF EXISTS "projects";
DROP TABLE IF EXISTS "issues";
DROP TABLE IF EXISTS "issue_comments";
DROP TABLE IF EXISTS "heartbeat_runs";
DROP TABLE IF EXISTS "goals";
DROP TABLE IF EXISTS "cost_events";
DROP TABLE IF EXISTS "companies";
DROP TABLE IF EXISTS "approvals";
DROP TABLE IF EXISTS "agents";
DROP TABLE IF EXISTS "agent_api_keys";
DROP TABLE IF EXISTS "activity_log";
