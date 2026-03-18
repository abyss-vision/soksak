-- +goose Up
ALTER TABLE "join_requests" ADD COLUMN "claim_secret_hash" text;
ALTER TABLE "join_requests" ADD COLUMN "claim_secret_expires_at" timestamp with time zone;
ALTER TABLE "join_requests" ADD COLUMN "claim_secret_consumed_at" timestamp with time zone;

-- +goose Down
ALTER TABLE "join_requests" DROP COLUMN IF EXISTS "claim_secret_consumed_at";
ALTER TABLE "join_requests" DROP COLUMN IF EXISTS "claim_secret_expires_at";
ALTER TABLE "join_requests" DROP COLUMN IF EXISTS "claim_secret_hash";
