-- +goose Up
ALTER TABLE "project_workspaces" ALTER COLUMN "cwd" DROP NOT NULL;

-- +goose Down
UPDATE "project_workspaces" SET "cwd" = '' WHERE "cwd" IS NULL;
ALTER TABLE "project_workspaces" ALTER COLUMN "cwd" SET NOT NULL;
