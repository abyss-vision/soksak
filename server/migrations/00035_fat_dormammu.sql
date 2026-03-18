-- +goose Up
DROP INDEX "budget_incidents_policy_window_threshold_idx";
CREATE UNIQUE INDEX "budget_incidents_policy_window_threshold_idx" ON "budget_incidents" USING btree ("policy_uuid","window_start","threshold_type") WHERE "budget_incidents"."status" <> 'dismissed';

-- +goose Down
DROP INDEX IF EXISTS "budget_incidents_policy_window_threshold_idx";
CREATE UNIQUE INDEX "budget_incidents_policy_window_threshold_idx" ON "budget_incidents" USING btree ("policy_uuid","window_start","threshold_type");
