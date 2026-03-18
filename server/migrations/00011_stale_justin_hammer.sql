-- +goose Up
CREATE TABLE "assets" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"provider" text NOT NULL,
	"object_key" text NOT NULL,
	"content_type" text NOT NULL,
	"byte_size" integer NOT NULL,
	"sha256" text NOT NULL,
	"original_filename" text,
	"created_by_agent_uuid" uuid,
	"created_by_user_id" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "issue_attachments" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"issue_uuid" uuid NOT NULL,
	"asset_uuid" uuid NOT NULL,
	"issue_comment_uuid" uuid,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "assets" ADD CONSTRAINT "assets_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "assets" ADD CONSTRAINT "assets_created_by_agent_uuid_fk" FOREIGN KEY ("created_by_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issue_attachments" ADD CONSTRAINT "issue_attachments_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issue_attachments" ADD CONSTRAINT "issue_attachments_issue_uuid_fk" FOREIGN KEY ("issue_uuid") REFERENCES "public"."issues"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "issue_attachments" ADD CONSTRAINT "issue_attachments_asset_uuid_fk" FOREIGN KEY ("asset_uuid") REFERENCES "public"."assets"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "issue_attachments" ADD CONSTRAINT "issue_attachments_issue_comment_uuid_fk" FOREIGN KEY ("issue_comment_uuid") REFERENCES "public"."issue_comments"("uuid") ON DELETE set null ON UPDATE no action;
CREATE INDEX "assets_company_created_idx" ON "assets" USING btree ("company_uuid","created_at");
CREATE INDEX "assets_company_provider_idx" ON "assets" USING btree ("company_uuid","provider");
CREATE UNIQUE INDEX "assets_company_object_key_uq" ON "assets" USING btree ("company_uuid","object_key");
CREATE INDEX "issue_attachments_company_issue_idx" ON "issue_attachments" USING btree ("company_uuid","issue_uuid");
CREATE INDEX "issue_attachments_issue_comment_idx" ON "issue_attachments" USING btree ("issue_comment_uuid");
CREATE UNIQUE INDEX "issue_attachments_asset_uq" ON "issue_attachments" USING btree ("asset_uuid");

-- +goose Down
DROP INDEX IF EXISTS "issue_attachments_asset_uq";
DROP INDEX IF EXISTS "issue_attachments_issue_comment_idx";
DROP INDEX IF EXISTS "issue_attachments_company_issue_idx";
DROP INDEX IF EXISTS "assets_company_object_key_uq";
DROP INDEX IF EXISTS "assets_company_provider_idx";
DROP INDEX IF EXISTS "assets_company_created_idx";
DROP TABLE IF EXISTS "issue_attachments";
DROP TABLE IF EXISTS "assets";
