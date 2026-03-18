-- +goose Up
CREATE TABLE "agent_config_revisions" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"agent_uuid" uuid NOT NULL,
	"created_by_agent_uuid" uuid,
	"created_by_user_id" text,
	"source" text DEFAULT 'patch' NOT NULL,
	"rolled_back_from_revision_uuid" uuid,
	"changed_keys" jsonb DEFAULT '[]'::jsonb NOT NULL,
	"before_config" jsonb NOT NULL,
	"after_config" jsonb NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "issue_approvals" (
	"company_uuid" uuid NOT NULL,
	"issue_uuid" uuid NOT NULL,
	"approval_uuid" uuid NOT NULL,
	"linked_by_agent_uuid" uuid,
	"linked_by_user_id" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	CONSTRAINT "issue_approvals_pk" PRIMARY KEY("issue_uuid","approval_uuid")
);
ALTER TABLE "agent_config_revisions" ADD CONSTRAINT "agent_config_revisions_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "agent_config_revisions" ADD CONSTRAINT "agent_config_revisions_agent_uuid_fk" FOREIGN KEY ("agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "agent_config_revisions" ADD CONSTRAINT "agent_config_revisions_created_by_agent_uuid_fk" FOREIGN KEY ("created_by_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "issue_approvals" ADD CONSTRAINT "issue_approvals_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issue_approvals" ADD CONSTRAINT "issue_approvals_issue_uuid_fk" FOREIGN KEY ("issue_uuid") REFERENCES "public"."issues"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "issue_approvals" ADD CONSTRAINT "issue_approvals_approval_uuid_fk" FOREIGN KEY ("approval_uuid") REFERENCES "public"."approvals"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "issue_approvals" ADD CONSTRAINT "issue_approvals_linked_by_agent_uuid_fk" FOREIGN KEY ("linked_by_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE set null ON UPDATE no action;
CREATE INDEX "agent_config_revisions_company_agent_created_idx" ON "agent_config_revisions" USING btree ("company_uuid","agent_uuid","created_at");
CREATE INDEX "agent_config_revisions_agent_created_idx" ON "agent_config_revisions" USING btree ("agent_uuid","created_at");
CREATE INDEX "issue_approvals_issue_idx" ON "issue_approvals" USING btree ("issue_uuid");
CREATE INDEX "issue_approvals_approval_idx" ON "issue_approvals" USING btree ("approval_uuid");
CREATE INDEX "issue_approvals_company_idx" ON "issue_approvals" USING btree ("company_uuid");

-- +goose Down
DROP INDEX IF EXISTS "issue_approvals_company_idx";
DROP INDEX IF EXISTS "issue_approvals_approval_idx";
DROP INDEX IF EXISTS "issue_approvals_issue_idx";
DROP INDEX IF EXISTS "agent_config_revisions_agent_created_idx";
DROP INDEX IF EXISTS "agent_config_revisions_company_agent_created_idx";
DROP TABLE IF EXISTS "issue_approvals";
DROP TABLE IF EXISTS "agent_config_revisions";
