-- +goose Up
CREATE TABLE "issue_labels" (
	"issue_uuid" uuid NOT NULL,
	"label_uuid" uuid NOT NULL,
	"company_uuid" uuid NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	CONSTRAINT "issue_labels_pk" PRIMARY KEY("issue_uuid","label_uuid")
);
CREATE TABLE "labels" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"name" text NOT NULL,
	"color" text NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "issue_labels" ADD CONSTRAINT "issue_labels_issue_uuid_fk" FOREIGN KEY ("issue_uuid") REFERENCES "public"."issues"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "issue_labels" ADD CONSTRAINT "issue_labels_label_uuid_fk" FOREIGN KEY ("label_uuid") REFERENCES "public"."labels"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "issue_labels" ADD CONSTRAINT "issue_labels_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "labels" ADD CONSTRAINT "labels_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE cascade ON UPDATE no action;
CREATE INDEX "issue_labels_issue_idx" ON "issue_labels" USING btree ("issue_uuid");
CREATE INDEX "issue_labels_label_idx" ON "issue_labels" USING btree ("label_uuid");
CREATE INDEX "issue_labels_company_idx" ON "issue_labels" USING btree ("company_uuid");
CREATE INDEX "labels_company_idx" ON "labels" USING btree ("company_uuid");
CREATE UNIQUE INDEX "labels_company_name_idx" ON "labels" USING btree ("company_uuid","name");

-- +goose Down
DROP INDEX IF EXISTS "labels_company_name_idx";
DROP INDEX IF EXISTS "labels_company_idx";
DROP INDEX IF EXISTS "issue_labels_company_idx";
DROP INDEX IF EXISTS "issue_labels_label_idx";
DROP INDEX IF EXISTS "issue_labels_issue_idx";
DROP TABLE IF EXISTS "labels";
DROP TABLE IF EXISTS "issue_labels";
