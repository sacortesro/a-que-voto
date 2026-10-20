-- =====================================================================
-- 13_rename_legislature.sql: English name for legislatura and its FK column
-- Depends on: 05, 08
-- =====================================================================

-- +goose up
-- +goose StatementBegin

ALTER TABLE legislatura RENAME TO legislature;
ALTER TRIGGER legislatura_audit ON legislature RENAME TO legislature_audit;
ALTER TABLE bill RENAME COLUMN legislatura_radicacion_id TO legislature_filing_id;

-- +goose StatementEnd

-- +goose down
-- +goose StatementBegin

ALTER TABLE bill RENAME COLUMN legislature_filing_id TO legislatura_radicacion_id;
ALTER TRIGGER legislature_audit ON legislature RENAME TO legislatura_audit;
ALTER TABLE legislature RENAME TO legislatura;

-- +goose StatementEnd
