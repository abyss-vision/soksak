-- +goose Up
CREATE INDEX "issue_comments_company_issue_created_at_idx" ON "issue_comments" USING btree ("company_uuid","issue_uuid","created_at");
CREATE INDEX "issue_comments_company_author_issue_created_at_idx" ON "issue_comments" USING btree ("company_uuid","author_user_id","issue_uuid","created_at");

-- +goose Down
DROP INDEX IF EXISTS "issue_comments_company_author_issue_created_at_idx";
DROP INDEX IF EXISTS "issue_comments_company_issue_created_at_idx";
