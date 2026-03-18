-- +goose Up
CREATE TABLE "board_settings" (
	"uuid" uuid PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
	"company_uuid" uuid NOT NULL,
	"column_order" jsonb NOT NULL DEFAULT '[]'::jsonb,
	"hidden_columns" jsonb NOT NULL DEFAULT '[]'::jsonb,
	"swim_lane_field" varchar(50) DEFAULT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL,
	"updated_at" timestamp with time zone DEFAULT now() NOT NULL
);
ALTER TABLE "board_settings" ADD CONSTRAINT "board_settings_company_uuid_fk" FOREIGN KEY ("company_uuid") REFERENCES "companies"("uuid") ON DELETE CASCADE ON UPDATE no action;
CREATE UNIQUE INDEX "board_settings_company_uuid_idx" ON "board_settings" ("company_uuid");

-- +goose Down
DROP TABLE IF EXISTS "board_settings";
