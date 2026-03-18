-- +goose Up
ALTER TABLE "companies" ADD COLUMN "issue_prefix" text NOT NULL DEFAULT 'PAP';
ALTER TABLE "companies" ADD COLUMN "issue_counter" integer NOT NULL DEFAULT 0;
ALTER TABLE "issues" ADD COLUMN "issue_number" integer;
ALTER TABLE "issues" ADD COLUMN "identifier" text;

WITH numbered AS (
  SELECT id, company_id, ROW_NUMBER() OVER (PARTITION BY company_id ORDER BY created_at ASC) AS rn
  FROM issues
)
UPDATE issues
SET issue_number = numbered.rn,
    identifier = (SELECT issue_prefix FROM companies WHERE companies.id = issues.company_id) || '-' || numbered.rn
FROM numbered
WHERE issues.id = numbered.id;

UPDATE companies
SET issue_counter = COALESCE(
  (SELECT MAX(issue_number) FROM issues WHERE issues.company_id = companies.id),
  0
);

CREATE UNIQUE INDEX "issues_company_identifier_idx" ON "issues" USING btree ("company_id","identifier");

-- +goose Down
DROP INDEX IF EXISTS "issues_company_identifier_idx";
ALTER TABLE "issues" DROP COLUMN IF EXISTS "identifier";
ALTER TABLE "issues" DROP COLUMN IF EXISTS "issue_number";
ALTER TABLE "companies" DROP COLUMN IF EXISTS "issue_counter";
ALTER TABLE "companies" DROP COLUMN IF EXISTS "issue_prefix";
