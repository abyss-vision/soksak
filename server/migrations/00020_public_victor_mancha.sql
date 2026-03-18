-- +goose Up
CREATE TABLE "project_workspaces" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"project_uuid" uuid NOT NULL,
	"name" text NOT NULL,
	"cwd" text NOT NULL,
	"repo_url" text,
	"repo_ref" text,
	"metadata" jsonb,
	"is_primary" boolean DEFAULT false NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "project_workspaces" ADD CONSTRAINT "project_workspaces_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "project_workspaces" ADD CONSTRAINT "project_workspaces_project_uuid_fk" FOREIGN KEY ("project_uuid") REFERENCES "public"."projects"("uuid") ON DELETE cascade ON UPDATE no action;
CREATE INDEX "project_workspaces_company_project_idx" ON "project_workspaces" USING btree ("company_uuid","project_uuid");
CREATE INDEX "project_workspaces_project_primary_idx" ON "project_workspaces" USING btree ("project_uuid","is_primary");

-- +goose Down
DROP INDEX IF EXISTS "project_workspaces_project_primary_idx";
DROP INDEX IF EXISTS "project_workspaces_company_project_idx";
DROP TABLE IF EXISTS "project_workspaces";
