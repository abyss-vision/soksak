-- +goose Up
ALTER TABLE "companies" ADD COLUMN "brand_color" text;

-- +goose Down
ALTER TABLE "companies" DROP COLUMN IF EXISTS "brand_color";
