-- +goose Up
CREATE TABLE "workspace_operations" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"execution_workspace_uuid" uuid,
	"heartbeat_run_uuid" uuid,
	"phase" text NOT NULL,
	"command" text,
	"cwd" text,
	"status" text DEFAULT 'running' NOT NULL,
	"exit_code" integer,
	"log_store" text,
	"log_ref" text,
	"log_bytes" bigint,
	"log_sha256" text,
	"log_compressed" boolean DEFAULT false NOT NULL,
	"stdout_excerpt" text,
	"stderr_excerpt" text,
	"metadata" jsonb,
	"started_at" timestamp with time zone DEFAULT now() NOT NULL,
	"finished_at" timestamp with time zone,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "workspace_operations" ADD CONSTRAINT "workspace_operations_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "workspace_operations" ADD CONSTRAINT "workspace_operations_execution_workspace_uuid_fk" FOREIGN KEY ("execution_workspace_uuid") REFERENCES "public"."execution_workspaces"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "workspace_operations" ADD CONSTRAINT "workspace_operations_heartbeat_run_uuid_fk" FOREIGN KEY ("heartbeat_run_uuid") REFERENCES "public"."heartbeat_runs"("uuid") ON DELETE set null ON UPDATE no action;
CREATE INDEX "workspace_operations_company_run_started_idx" ON "workspace_operations" USING btree ("company_uuid","heartbeat_run_uuid","started_at");
CREATE INDEX "workspace_operations_company_workspace_started_idx" ON "workspace_operations" USING btree ("company_uuid","execution_workspace_uuid","started_at");

-- +goose Down
DROP TABLE IF EXISTS "workspace_operations";
