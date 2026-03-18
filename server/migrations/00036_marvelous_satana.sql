-- +goose Up
CREATE TABLE "execution_workspaces" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"project_uuid" uuid NOT NULL,
	"project_workspace_uuid" uuid,
	"source_issue_uuid" uuid,
	"mode" text NOT NULL,
	"strategy_type" text NOT NULL,
	"name" text NOT NULL,
	"status" text DEFAULT 'active' NOT NULL,
	"cwd" text,
	"repo_url" text,
	"base_ref" text,
	"branch_name" text,
	"provider_type" text DEFAULT 'local_fs' NOT NULL,
	"provider_ref" text,
	"derived_from_execution_workspace_uuid" uuid,
	"last_used_at" timestamp with time zone DEFAULT now() NOT NULL,
	"opened_at" timestamp with time zone DEFAULT now() NOT NULL,
	"closed_at" timestamp with time zone,
	"cleanup_eligible_at" timestamp with time zone,
	"cleanup_reason" text,
	"metadata" jsonb,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "issue_work_products" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"project_uuid" uuid,
	"issue_uuid" uuid NOT NULL,
	"execution_workspace_uuid" uuid,
	"runtime_service_uuid" uuid,
	"type" text NOT NULL,
	"provider" text NOT NULL,
	"external_id" text,
	"title" text NOT NULL,
	"url" text,
	"status" text NOT NULL,
	"review_state" text DEFAULT 'none' NOT NULL,
	"is_primary" boolean DEFAULT false NOT NULL,
	"health_status" text DEFAULT 'unknown' NOT NULL,
	"summary" text,
	"metadata" jsonb,
	"created_by_run_uuid" uuid,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "issues" ADD COLUMN "project_workspace_uuid" uuid;
ALTER TABLE "issues" ADD COLUMN "execution_workspace_uuid" uuid;
ALTER TABLE "issues" ADD COLUMN "execution_workspace_preference" text;
ALTER TABLE "project_workspaces" ADD COLUMN "source_type" text DEFAULT 'local_path' NOT NULL;
ALTER TABLE "project_workspaces" ADD COLUMN "default_ref" text;
ALTER TABLE "project_workspaces" ADD COLUMN "visibility" text DEFAULT 'default' NOT NULL;
ALTER TABLE "project_workspaces" ADD COLUMN "setup_command" text;
ALTER TABLE "project_workspaces" ADD COLUMN "cleanup_command" text;
ALTER TABLE "project_workspaces" ADD COLUMN "remote_provider" text;
ALTER TABLE "project_workspaces" ADD COLUMN "remote_workspace_ref" text;
ALTER TABLE "project_workspaces" ADD COLUMN "shared_workspace_key" text;
ALTER TABLE "workspace_runtime_services" ADD COLUMN "execution_workspace_uuid" uuid;
ALTER TABLE "execution_workspaces" ADD CONSTRAINT "execution_workspaces_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "execution_workspaces" ADD CONSTRAINT "execution_workspaces_project_uuid_fk" FOREIGN KEY ("project_uuid") REFERENCES "public"."projects"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "execution_workspaces" ADD CONSTRAINT "execution_workspaces_project_workspace_uuid_fk" FOREIGN KEY ("project_workspace_uuid") REFERENCES "public"."project_workspaces"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "execution_workspaces" ADD CONSTRAINT "execution_workspaces_source_issue_uuid_fk" FOREIGN KEY ("source_issue_uuid") REFERENCES "public"."issues"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "execution_workspaces" ADD CONSTRAINT "execution_workspaces_derived_from_execution_workspace_uuid_fk" FOREIGN KEY ("derived_from_execution_workspace_uuid") REFERENCES "public"."execution_workspaces"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "issue_work_products" ADD CONSTRAINT "issue_work_products_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issue_work_products" ADD CONSTRAINT "issue_work_products_project_uuid_fk" FOREIGN KEY ("project_uuid") REFERENCES "public"."projects"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "issue_work_products" ADD CONSTRAINT "issue_work_products_issue_uuid_fk" FOREIGN KEY ("issue_uuid") REFERENCES "public"."issues"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "issue_work_products" ADD CONSTRAINT "issue_work_products_execution_workspace_uuid_fk" FOREIGN KEY ("execution_workspace_uuid") REFERENCES "public"."execution_workspaces"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "issue_work_products" ADD CONSTRAINT "issue_work_products_runtime_service_uuid_fk" FOREIGN KEY ("runtime_service_uuid") REFERENCES "public"."workspace_runtime_services"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "issue_work_products" ADD CONSTRAINT "issue_work_products_created_by_run_uuid_fk" FOREIGN KEY ("created_by_run_uuid") REFERENCES "public"."heartbeat_runs"("uuid") ON DELETE set null ON UPDATE no action;
CREATE INDEX "execution_workspaces_company_project_status_idx" ON "execution_workspaces" USING btree ("company_uuid","project_uuid","status");
CREATE INDEX "execution_workspaces_company_project_workspace_status_idx" ON "execution_workspaces" USING btree ("company_uuid","project_workspace_uuid","status");
CREATE INDEX "execution_workspaces_company_source_issue_idx" ON "execution_workspaces" USING btree ("company_uuid","source_issue_uuid");
CREATE INDEX "execution_workspaces_company_last_used_idx" ON "execution_workspaces" USING btree ("company_uuid","last_used_at");
CREATE INDEX "execution_workspaces_company_branch_idx" ON "execution_workspaces" USING btree ("company_uuid","branch_name");
CREATE INDEX "issue_work_products_company_issue_type_idx" ON "issue_work_products" USING btree ("company_uuid","issue_uuid","type");
CREATE INDEX "issue_work_products_company_execution_workspace_type_idx" ON "issue_work_products" USING btree ("company_uuid","execution_workspace_uuid","type");
CREATE INDEX "issue_work_products_company_provider_external_id_idx" ON "issue_work_products" USING btree ("company_uuid","provider","external_id");
CREATE INDEX "issue_work_products_company_updated_idx" ON "issue_work_products" USING btree ("company_uuid","updated_at");
ALTER TABLE "issues" ADD CONSTRAINT "issues_project_workspace_uuid_fk" FOREIGN KEY ("project_workspace_uuid") REFERENCES "public"."project_workspaces"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "issues" ADD CONSTRAINT "issues_execution_workspace_uuid_fk" FOREIGN KEY ("execution_workspace_uuid") REFERENCES "public"."execution_workspaces"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "workspace_runtime_services" ADD CONSTRAINT "workspace_runtime_services_execution_workspace_uuid_fk" FOREIGN KEY ("execution_workspace_uuid") REFERENCES "public"."execution_workspaces"("uuid") ON DELETE set null ON UPDATE no action;
CREATE INDEX "issues_company_project_workspace_idx" ON "issues" USING btree ("company_uuid","project_workspace_uuid");
CREATE INDEX "issues_company_execution_workspace_idx" ON "issues" USING btree ("company_uuid","execution_workspace_uuid");
CREATE INDEX "project_workspaces_project_source_type_idx" ON "project_workspaces" USING btree ("project_uuid","source_type");
CREATE INDEX "project_workspaces_company_shared_key_idx" ON "project_workspaces" USING btree ("company_uuid","shared_workspace_key");
CREATE UNIQUE INDEX "project_workspaces_project_remote_ref_idx" ON "project_workspaces" USING btree ("project_uuid","remote_provider","remote_workspace_ref");
CREATE INDEX "workspace_runtime_services_company_execution_workspace_status_idx" ON "workspace_runtime_services" USING btree ("company_uuid","execution_workspace_uuid","status");

-- +goose Down
DROP TABLE IF EXISTS "issue_work_products";
DROP TABLE IF EXISTS "execution_workspaces";
ALTER TABLE "issues" DROP COLUMN IF EXISTS "execution_workspace_preference";
ALTER TABLE "issues" DROP COLUMN IF EXISTS "execution_workspace_uuid";
ALTER TABLE "issues" DROP COLUMN IF EXISTS "project_workspace_uuid";
ALTER TABLE "project_workspaces" DROP COLUMN IF EXISTS "shared_workspace_key";
ALTER TABLE "project_workspaces" DROP COLUMN IF EXISTS "remote_workspace_ref";
ALTER TABLE "project_workspaces" DROP COLUMN IF EXISTS "remote_provider";
ALTER TABLE "project_workspaces" DROP COLUMN IF EXISTS "cleanup_command";
ALTER TABLE "project_workspaces" DROP COLUMN IF EXISTS "setup_command";
ALTER TABLE "project_workspaces" DROP COLUMN IF EXISTS "visibility";
ALTER TABLE "project_workspaces" DROP COLUMN IF EXISTS "default_ref";
ALTER TABLE "project_workspaces" DROP COLUMN IF EXISTS "source_type";
ALTER TABLE "workspace_runtime_services" DROP COLUMN IF EXISTS "execution_workspace_uuid";
