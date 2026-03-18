-- +goose Up
CREATE TABLE "company_logos" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"asset_uuid" uuid NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "company_logos" ADD CONSTRAINT "company_logos_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "public"."companies"("uuid") ON DELETE cascade ON UPDATE no action;
ALTER TABLE "company_logos" ADD CONSTRAINT "company_logos_asset_uuid_fk" FOREIGN KEY ("asset_uuid") REFERENCES "public"."assets"("uuid") ON DELETE cascade ON UPDATE no action;
CREATE UNIQUE INDEX "company_logos_company_uq" ON "company_logos" USING btree ("company_uuid");
CREATE UNIQUE INDEX "company_logos_asset_uq" ON "company_logos" USING btree ("asset_uuid");

-- +goose Down
DROP TABLE IF EXISTS "company_logos";
