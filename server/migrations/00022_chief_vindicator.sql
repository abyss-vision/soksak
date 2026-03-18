-- +goose Up
ALTER TABLE "issues" ADD COLUMN "assignee_adapter_overrides" jsonb;

-- +goose Down
ALTER TABLE "issues" DROP COLUMN IF EXISTS "assignee_adapter_overrides";
