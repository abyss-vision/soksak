-- +goose Up
CREATE TABLE "document_revisions" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"document_uuid" uuid NOT NULL,
	"revision_number" integer NOT NULL,
	"body" text NOT NULL,
	"change_summary" text,
	"created_by_agent_uuid" uuid,
	"created_by_user_id" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "documents" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"title" text,
	"format" text DEFAULT 'markdown' NOT NULL,
	"latest_body" text NOT NULL,
	"latest_revision_uuid" uuid,
	"latest_revision_number" integer DEFAULT 1 NOT NULL,
	"created_by_agent_uuid" uuid,
	"created_by_user_id" text,
	"updated_by_agent_uuid" uuid,
	"updated_by_user_id" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "issue_documents" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"issue_uuid" uuid NOT NULL,
	"document_uuid" uuid NOT NULL,
	"key" text NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "document_revisions" ADD CONSTRAINT "document_revisions_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "document_revisions" ADD CONSTRAINT "document_revisions_document_uuid_fk" FOREIGN KEY ("document_uuid") REFERENCES "public"."documents"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "document_revisions" ADD CONSTRAINT "document_revisions_created_by_agent_uuid_fk" FOREIGN KEY ("created_by_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "documents" ADD CONSTRAINT "documents_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "documents" ADD CONSTRAINT "documents_created_by_agent_uuid_fk" FOREIGN KEY ("created_by_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "documents" ADD CONSTRAINT "documents_updated_by_agent_uuid_fk" FOREIGN KEY ("updated_by_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "issue_documents" ADD CONSTRAINT "issue_documents_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issue_documents" ADD CONSTRAINT "issue_documents_issue_uuid_fk" FOREIGN KEY ("issue_uuid") REFERENCES "public"."issues"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "issue_documents" ADD CONSTRAINT "issue_documents_document_uuid_fk" FOREIGN KEY ("document_uuid") REFERENCES "public"."documents"("uuid") ON DELETE cascade ON UPDATE no action;
CREATE UNIQUE INDEX "document_revisions_document_revision_uq" ON "document_revisions" USING btree ("document_uuid","revision_number");
CREATE INDEX "document_revisions_company_document_created_idx" ON "document_revisions" USING btree ("company_uuid","document_uuid","created_at");
CREATE INDEX "documents_company_updated_idx" ON "documents" USING btree ("company_uuid","updated_at");
CREATE INDEX "documents_company_created_idx" ON "documents" USING btree ("company_uuid","created_at");
CREATE UNIQUE INDEX "issue_documents_company_issue_key_uq" ON "issue_documents" USING btree ("company_uuid","issue_uuid","key");
CREATE UNIQUE INDEX "issue_documents_document_uq" ON "issue_documents" USING btree ("document_uuid");
CREATE INDEX "issue_documents_company_issue_updated_idx" ON "issue_documents" USING btree ("company_uuid","issue_uuid","updated_at");

-- +goose Down
DROP TABLE IF EXISTS "issue_documents";
DROP TABLE IF EXISTS "documents";
DROP TABLE IF EXISTS "document_revisions";
