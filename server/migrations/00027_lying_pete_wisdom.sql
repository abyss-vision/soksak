-- +goose Up
CREATE TABLE "workspace_runtime_services" (
	"uuid" uuid PRIMARY KEY NOT NULL,
	"company_uuid" uuid NOT NULL,
	"project_uuid" uuid,
	"project_workspace_uuid" uuid,
	"issue_uuid" uuid,
	"scope_type" text NOT NULL,
	"scope_id" text,
	"service_name" text NOT NULL,
	"status" text NOT NULL,
	"lifecycle" text NOT NULL,
	"reuse_key" text,
	"command" text,
	"cwd" text,
	"port" integer,
	"url" text,
	"provider" text NOT NULL,
	"provider_ref" text,
	"owner_agent_uuid" uuid,
	"started_by_run_uuid" uuid,
	"last_used_at" timestamp with time zone DEFAULT now() NOT NULL,
	"started_at" timestamp with time zone DEFAULT now() NOT NULL,
	"stopped_at" timestamp with time zone,
	"stop_policy" jsonb,
	"health_status" text DEFAULT 'unknown' NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "workspace_runtime_services" ADD CONSTRAINT "workspace_runtime_services_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "workspace_runtime_services" ADD CONSTRAINT "workspace_runtime_services_project_uuid_fk" FOREIGN KEY ("project_uuid") REFERENCES "public"."projects"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "workspace_runtime_services" ADD CONSTRAINT "workspace_runtime_services_project_workspace_uuid_fk" FOREIGN KEY ("project_workspace_uuid") REFERENCES "public"."project_workspaces"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "workspace_runtime_services" ADD CONSTRAINT "workspace_runtime_services_issue_uuid_fk" FOREIGN KEY ("issue_uuid") REFERENCES "public"."issues"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "workspace_runtime_services" ADD CONSTRAINT "workspace_runtime_services_owner_agent_uuid_fk" FOREIGN KEY ("owner_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "workspace_runtime_services" ADD CONSTRAINT "workspace_runtime_services_started_by_run_uuid_fk" FOREIGN KEY ("started_by_run_uuid") REFERENCES "public"."heartbeat_runs"("uuid") ON DELETE set null ON UPDATE no action;
CREATE INDEX "workspace_runtime_services_company_workspace_status_idx" ON "workspace_runtime_services" USING btree ("company_uuid","project_workspace_uuid","status");
CREATE INDEX "workspace_runtime_services_company_project_status_idx" ON "workspace_runtime_services" USING btree ("company_uuid","project_uuid","status");
CREATE INDEX "workspace_runtime_services_run_idx" ON "workspace_runtime_services" USING btree ("started_by_run_uuid");
CREATE INDEX "workspace_runtime_services_company_updated_idx" ON "workspace_runtime_services" USING btree ("company_uuid","updated_at");

-- +goose Down
DROP TABLE IF EXISTS "workspace_runtime_services";
