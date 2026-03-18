-- +goose Up
DROP INDEX IF EXISTS "issues_company_identifier_idx";

WITH ranked_companies AS (
  SELECT
    c.uuid,
    COALESCE(NULLIF(SUBSTRING(REGEXP_REPLACE(UPPER(c.name), '[^A-Z]', '', 'g') FROM 1 FOR 3), ''), 'CMP') AS base_prefix,
    ROW_NUMBER() OVER (
      PARTITION BY COALESCE(NULLIF(SUBSTRING(REGEXP_REPLACE(UPPER(c.name), '[^A-Z]', '', 'g') FROM 1 FOR 3), ''), 'CMP')
      ORDER BY c.created_at, c.uuid
    ) AS prefix_rank
  FROM companies c
)
UPDATE companies c
SET issue_prefix = CASE
  WHEN ranked_companies.prefix_rank = 1 THEN ranked_companies.base_prefix
  ELSE ranked_companies.base_prefix || REPEAT('A', (ranked_companies.prefix_rank - 1)::integer)
END
FROM ranked_companies
WHERE c.uuid = ranked_companies.uuid;

WITH numbered_issues AS (
  SELECT
    i.uuid,
    ROW_NUMBER() OVER (PARTITION BY i.company_uuid ORDER BY i.created_at, i.uuid) AS issue_number
  FROM issues i
)
UPDATE issues i
SET issue_number = numbered_issues.issue_number
FROM numbered_issues
WHERE i.uuid = numbered_issues.uuid;

UPDATE issues i
SET identifier = c.issue_prefix || '-' || i.issue_number
FROM companies c
WHERE c.uuid = i.company_uuid;

UPDATE companies c
SET issue_counter = COALESCE((
  SELECT MAX(i.issue_number)
  FROM issues i
  WHERE i.company_uuid = c.uuid
), 0);

CREATE UNIQUE INDEX "companies_issue_prefix_idx" ON "companies" USING btree ("issue_prefix");
CREATE UNIQUE INDEX "issues_identifier_idx" ON "issues" USING btree ("identifier");

-- +goose Down
DROP INDEX IF EXISTS "issues_identifier_idx";
DROP INDEX IF EXISTS "companies_issue_prefix_idx";
