-- +goose Up
CREATE TABLE "approval_comments" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"approval_uuid" uuid NOT NULL,
	"author_agent_uuid" uuid,
	"author_user_id" text,
	"body" text NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "agents" ADD COLUMN "permissions" jsonb DEFAULT '{}'::jsonb NOT NULL;
ALTER TABLE "companies" ADD COLUMN "require_board_approval_for_new_agents" boolean DEFAULT true NOT NULL;
ALTER TABLE "approval_comments" ADD CONSTRAINT "approval_comments_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "approval_comments" ADD CONSTRAINT "approval_comments_approval_uuid_fk" FOREIGN KEY ("approval_uuid") REFERENCES "public"."approvals"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "approval_comments" ADD CONSTRAINT "approval_comments_author_agent_uuid_fk" FOREIGN KEY ("author_agent_uuid") REFERENCES "public"."agents"("uuid") ON DELETE no action ON UPDATE no action;
CREATE INDEX "approval_comments_company_idx" ON "approval_comments" USING btree ("company_uuid");
CREATE INDEX "approval_comments_approval_idx" ON "approval_comments" USING btree ("approval_uuid");
CREATE INDEX "approval_comments_approval_created_idx" ON "approval_comments" USING btree ("approval_uuid","created_at");

-- +goose Down
DROP INDEX IF EXISTS "approval_comments_approval_created_idx";
DROP INDEX IF EXISTS "approval_comments_approval_idx";
DROP INDEX IF EXISTS "approval_comments_company_idx";
DROP TABLE IF EXISTS "approval_comments";
ALTER TABLE "companies" DROP COLUMN IF EXISTS "require_board_approval_for_new_agents";
ALTER TABLE "agents" DROP COLUMN IF EXISTS "permissions";
