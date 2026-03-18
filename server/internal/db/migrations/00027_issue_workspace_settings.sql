-- +goose Up
ALTER TABLE "issues" ADD COLUMN "execution_workspace_settings" jsonb;
ALTER TABLE "projects" ADD COLUMN "execution_workspace_policy" jsonb;

-- +goose Down
ALTER TABLE "projects" DROP COLUMN IF EXISTS "execution_workspace_policy";
ALTER TABLE "issues" DROP COLUMN IF EXISTS "execution_workspace_settings";
