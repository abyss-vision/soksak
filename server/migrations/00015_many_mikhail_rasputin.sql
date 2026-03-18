-- +goose Up
CREATE TABLE "account" (
	"id" text PRIMARY KEY NOT NULL,
	"account_id" text NOT NULL,
	"provider_id" text NOT NULL,
	"user_id" text NOT NULL,
	"access_token" text,
	"refresh_token" text,
	"id_token" text,
	"access_token_expires_at" timestamp with time zone,
	"refresh_token_expires_at" timestamp with time zone,
	"scope" text,
	"password" text,
	"created_at" timestamp with time zone NOT NULL,
	"updated_at" timestamp with time zone NOT NULL
);
CREATE TABLE "session" (
	"id" text PRIMARY KEY NOT NULL,
	"expires_at" timestamp with time zone NOT NULL,
	"token" text NOT NULL,
	"created_at" timestamp with time zone NOT NULL,
	"updated_at" timestamp with time zone NOT NULL,
	"ip_address" text,
	"user_agent" text,
	"user_id" text NOT NULL
);
CREATE TABLE "user" (
	"id" text PRIMARY KEY NOT NULL,
	"name" text NOT NULL,
	"email" text NOT NULL,
	"email_verified" boolean DEFAULT false NOT NULL,
	"image" text,
	"created_at" timestamp with time zone NOT NULL,
	"updated_at" timestamp with time zone NOT NULL
);
CREATE TABLE "verification" (
	"id" text PRIMARY KEY NOT NULL,
	"identifier" text NOT NULL,
	"value" text NOT NULL,
	"expires_at" timestamp with time zone NOT NULL,
	"created_at" timestamp with time zone,
	"updated_at" timestamp with time zone
);
CREATE TABLE "company_memberships" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"principal_type" text NOT NULL,
	"principal_id" text NOT NULL,
	"status" text DEFAULT 'active' NOT NULL,
	"membership_role" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "instance_user_roles" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"user_id" text NOT NULL,
	"role" text DEFAULT 'instance_admin' NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "invites" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid,
	"invite_type" text DEFAULT 'company_join' NOT NULL,
	"token_hash" text NOT NULL,
	"allowed_join_types" text DEFAULT 'both' NOT NULL,
	"defaults_payload" jsonb,
	"expires_at" timestamp with time zone NOT NULL,
	"invited_by_user_id" text,
	"revoked_at" timestamp with time zone,
	"accepted_at" timestamp with time zone,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "join_requests" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"invite_uuid" uuid NOT NULL,
	"company_uuid" uuid NOT NULL,
	"request_type" text NOT NULL,
	"status" text DEFAULT 'pending_approval' NOT NULL,
	"request_ip" text NOT NULL,
	"requesting_user_id" text,
	"request_email_snapshot" text,
	"agent_name" text,
	"adapter_type" text,
	"capabilities" text,
	"agent_defaults_payload" jsonb,
	"created_agent_uuid" uuid,
	"approved_by_user_id" text,
	"approved_at" timestamp with time zone,
	"rejected_by_user_id" text,
	"rejected_at" timestamp with time zone,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
CREATE TABLE "principal_permission_grants" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"principal_type" text NOT NULL,
	"principal_id" text NOT NULL,
	"permission_key" text NOT NULL,
	"scope" jsonb,
	"granted_by_user_id" text,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "issues" ADD COLUMN "assignee_user_id" text;
ALTER TABLE "account" ADD CONSTRAINT "account_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "session" ADD CONSTRAINT "session_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "company_memberships" ADD CONSTRAINT "company_memberships_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "invites" ADD CONSTRAINT "invites_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "join_requests" ADD CONSTRAINT "join_requests_invite_uuid_fk" FOREIGN KEY ("invite_uuid") REFERENCES "public"."invites"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "join_requests" ADD CONSTRAINT "join_requests_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "join_requests" ADD CONSTRAINT "join_requests_created_agent_uuid_fk" FOREIGN KEY ("created_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "principal_permission_grants" ADD CONSTRAINT "principal_permission_grants_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
CREATE UNIQUE INDEX "company_memberships_company_principal_unique_idx" ON "company_memberships" USING btree ("company_uuid","principal_type","principal_id");
CREATE INDEX "company_memberships_principal_status_idx" ON "company_memberships" USING btree ("principal_type","principal_id","status");
CREATE INDEX "company_memberships_company_status_idx" ON "company_memberships" USING btree ("company_uuid","status");
CREATE UNIQUE INDEX "instance_user_roles_user_role_unique_idx" ON "instance_user_roles" USING btree ("user_id","role");
CREATE INDEX "instance_user_roles_role_idx" ON "instance_user_roles" USING btree ("role");
CREATE UNIQUE INDEX "invites_token_hash_unique_idx" ON "invites" USING btree ("token_hash");
CREATE INDEX "invites_company_invite_state_idx" ON "invites" USING btree ("company_uuid","invite_type","revoked_at","expires_at");
CREATE UNIQUE INDEX "join_requests_invite_unique_idx" ON "join_requests" USING btree ("invite_uuid");
CREATE INDEX "join_requests_company_status_type_created_idx" ON "join_requests" USING btree ("company_uuid","status","request_type","created_at");
CREATE UNIQUE INDEX "principal_permission_grants_unique_idx" ON "principal_permission_grants" USING btree ("company_uuid","principal_type","principal_id","permission_key");
CREATE INDEX "principal_permission_grants_company_permission_idx" ON "principal_permission_grants" USING btree ("company_uuid","permission_key");
CREATE INDEX "issues_company_assignee_user_status_idx" ON "issues" USING btree ("company_uuid","assignee_user_id","status");

-- +goose Down
DROP INDEX IF EXISTS "issues_company_assignee_user_status_idx";
DROP INDEX IF EXISTS "principal_permission_grants_company_permission_idx";
DROP INDEX IF EXISTS "principal_permission_grants_unique_idx";
DROP INDEX IF EXISTS "join_requests_company_status_type_created_idx";
DROP INDEX IF EXISTS "join_requests_invite_unique_idx";
DROP INDEX IF EXISTS "invites_company_invite_state_idx";
DROP INDEX IF EXISTS "invites_token_hash_unique_idx";
DROP INDEX IF EXISTS "instance_user_roles_role_idx";
DROP INDEX IF EXISTS "instance_user_roles_user_role_unique_idx";
DROP INDEX IF EXISTS "company_memberships_company_status_idx";
DROP INDEX IF EXISTS "company_memberships_principal_status_idx";
DROP INDEX IF EXISTS "company_memberships_company_principal_unique_idx";
ALTER TABLE "issues" DROP COLUMN IF EXISTS "assignee_user_id";
DROP TABLE IF EXISTS "principal_permission_grants";
DROP TABLE IF EXISTS "join_requests";
DROP TABLE IF EXISTS "invites";
DROP TABLE IF EXISTS "instance_user_roles";
DROP TABLE IF EXISTS "company_memberships";
DROP TABLE IF EXISTS "verification";
DROP TABLE IF EXISTS "user";
DROP TABLE IF EXISTS "session";
DROP TABLE IF EXISTS "account";
