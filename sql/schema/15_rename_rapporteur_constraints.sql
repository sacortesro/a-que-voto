-- =====================================================================
-- 15_rename_rapporteur_constraints: English names for bill_ponente constraints and indexes
-- Depends on: 14
-- =====================================================================

-- +goose up
-- +goose StatementBegin

ALTER TABLE bill_rapporteur RENAME CONSTRAINT bill_ponente_pkey TO bill_rapporteur_pkey;
ALTER TABLE bill_rapporteur RENAME CONSTRAINT bill_ponente_bill_id_fkey TO bill_rapporteur_bill_id_fkey;
ALTER TABLE bill_rapporteur RENAME CONSTRAINT bill_ponente_mandate_id_fkey TO bill_rapporteur_mandate_id_fkey;
ALTER TABLE bill_rapporteur RENAME CONSTRAINT bill_ponente_stage_fkey TO bill_rapporteur_stage_fkey;
ALTER TABLE bill_rapporteur RENAME CONSTRAINT bill_ponente_committee_id_fkey TO bill_rapporteur_committee_id_fkey;
ALTER TABLE bill_rapporteur RENAME CONSTRAINT bill_ponente_created_by_fkey TO bill_rapporteur_created_by_fkey;
ALTER INDEX bill_ponente_mandate_id_idx RENAME TO bill_rapporteur_mandate_id_idx;

-- +goose StatementEnd

-- +goose down
-- +goose StatementBegin

ALTER INDEX bill_rapporteur_mandate_id_idx RENAME TO bill_ponente_mandate_id_idx;
ALTER TABLE bill_rapporteur RENAME CONSTRAINT bill_rapporteur_created_by_fkey TO bill_ponente_created_by_fkey;
ALTER TABLE bill_rapporteur RENAME CONSTRAINT bill_rapporteur_committee_id_fkey TO bill_ponente_committee_id_fkey;
ALTER TABLE bill_rapporteur RENAME CONSTRAINT bill_rapporteur_stage_fkey TO bill_ponente_stage_fkey;
ALTER TABLE bill_rapporteur RENAME CONSTRAINT bill_rapporteur_mandate_id_fkey TO bill_ponente_mandate_id_fkey;
ALTER TABLE bill_rapporteur RENAME CONSTRAINT bill_rapporteur_bill_id_fkey TO bill_ponente_bill_id_fkey;
ALTER TABLE bill_rapporteur RENAME CONSTRAINT bill_rapporteur_pkey TO bill_ponente_pkey;

-- +goose StatementEnd
