-- +goose Up
CREATE TABLE "project_goals" (
	"project_uuid" uuid NOT NULL,
	"goal_uuid" uuid NOT NULL,
	"company_uuid" uuid NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL,
	CONSTRAINT "project_goals_pk" PRIMARY KEY("project_uuid","goal_uuid")
);
ALTER TABLE "project_goals" ADD CONSTRAINT "project_goals_project_uuid_fk" FOREIGN KEY ("project_uuid") REFERENCES "public"."projects"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "project_goals" ADD CONSTRAINT "project_goals_goal_uuid_fk" FOREIGN KEY ("goal_uuid") REFERENCES "public"."goals"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "project_goals" ADD CONSTRAINT "project_goals_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE no action ON UPDATE no action;
CREATE INDEX "project_goals_project_idx" ON "project_goals" USING btree ("project_uuid");
CREATE INDEX "project_goals_goal_idx" ON "project_goals" USING btree ("goal_uuid");
CREATE INDEX "project_goals_company_idx" ON "project_goals" USING btree ("company_uuid");
INSERT INTO "project_goals" ("project_uuid", "goal_uuid", "company_uuid")
SELECT "uuid", "goal_uuid", "company_uuid" FROM "projects" WHERE "goal_uuid" IS NOT NULL
ON CONFLICT DO NOTHING;

-- +goose Down
DROP INDEX IF EXISTS "project_goals_company_idx";
DROP INDEX IF EXISTS "project_goals_goal_idx";
DROP INDEX IF EXISTS "project_goals_project_idx";
DROP TABLE IF EXISTS "project_goals";
