-- +goose Up
CREATE TABLE "issue_read_states" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"issue_uuid" uuid NOT NULL,
	"user_id" text NOT NULL,
	"last_read_at" timestamp with time zone DEFAULT now() NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "issue_read_states" ADD CONSTRAINT "issue_read_states_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
ALTER TABLE "issue_read_states" ADD CONSTRAINT "issue_read_states_issue_uuid_fk" FOREIGN KEY ("issue_uuid") REFERENCES "public"."issues"("uuid") ON DELETE no action ON UPDATE no action;
CREATE INDEX "issue_read_states_company_issue_idx" ON "issue_read_states" USING btree ("company_uuid","issue_uuid");
CREATE INDEX "issue_read_states_company_user_idx" ON "issue_read_states" USING btree ("company_uuid","user_id");
CREATE UNIQUE INDEX "issue_read_states_company_issue_user_idx" ON "issue_read_states" USING btree ("company_uuid","issue_uuid","user_id");

-- +goose Down
DROP INDEX IF EXISTS "issue_read_states_company_issue_user_idx";
DROP INDEX IF EXISTS "issue_read_states_company_user_idx";
DROP INDEX IF EXISTS "issue_read_states_company_issue_idx";
DROP TABLE IF EXISTS "issue_read_states";
