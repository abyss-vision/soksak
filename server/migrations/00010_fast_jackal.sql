-- +goose Up
CREATE TABLE "company_secret_versions" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"secret_uuid" uuid NOT NULL,
	"version" integer NOT NULL,
	"material" jsonb NOT NULL,
	"value_sha256" text NOT NULL,
	"created_by_agent_uuid" uuid,
	"created_by_user_id" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"revoked_at" timestamp with time zone
);
CREATE TABLE "company_secrets" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"name" text NOT NULL,
	"provider" text DEFAULT 'local_encrypted' NOT NULL,
	"external_ref" text,
	"latest_version" integer DEFAULT 1 NOT NULL,
	"description" text,
	"created_by_agent_uuid" uuid,
	"created_by_user_id" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "company_secret_versions" ADD CONSTRAINT "company_secret_versions_secret_uuid_fk" FOREIGN KEY ("secret_uuid") REFERENCES "public"."company_secrets"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "company_secret_versions" ADD CONSTRAINT "company_secret_versions_created_by_agent_uuid_fk" FOREIGN KEY ("created_by_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE set null ON UPDATE no action;
ALTER TABLE "company_secrets" ADD CONSTRAINT "company_secrets_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "company_secrets" ADD CONSTRAINT "company_secrets_created_by_agent_uuid_fk" FOREIGN KEY ("created_by_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE set null ON UPDATE no action;
CREATE INDEX "company_secret_versions_secret_idx" ON "company_secret_versions" USING btree ("secret_uuid","created_at");
CREATE INDEX "company_secret_versions_value_sha256_idx" ON "company_secret_versions" USING btree ("value_sha256");
CREATE UNIQUE INDEX "company_secret_versions_secret_version_uq" ON "company_secret_versions" USING btree ("secret_uuid","version");
CREATE INDEX "company_secrets_company_idx" ON "company_secrets" USING btree ("company_uuid");
CREATE INDEX "company_secrets_company_provider_idx" ON "company_secrets" USING btree ("company_uuid","provider");
CREATE UNIQUE INDEX "company_secrets_company_name_uq" ON "company_secrets" USING btree ("company_uuid","name");

-- +goose Down
DROP INDEX IF EXISTS "company_secrets_company_name_uq";
DROP INDEX IF EXISTS "company_secrets_company_provider_idx";
DROP INDEX IF EXISTS "company_secrets_company_idx";
DROP INDEX IF EXISTS "company_secret_versions_secret_version_uq";
DROP INDEX IF EXISTS "company_secret_versions_value_sha256_idx";
DROP INDEX IF EXISTS "company_secret_versions_secret_idx";
DROP TABLE IF EXISTS "company_secrets";
DROP TABLE IF EXISTS "company_secret_versions";
