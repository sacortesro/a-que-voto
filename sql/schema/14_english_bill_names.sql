-- =====================================================================
-- 14_english_bill_names.sql: English names for Spanish bill columns, indexes and table
-- Depends on: 08
-- =====================================================================

-- +goose up
-- +goose StatementBegin

ALTER TABLE bill RENAME COLUMN titulo TO title;
ALTER TABLE bill RENAME COLUMN titulo_norm TO title_norm;
ALTER TABLE bill RENAME COLUMN numero_camara TO chamber_number;
ALTER TABLE bill RENAME COLUMN numero_senado TO senate_number;
ALTER TABLE bill RENAME COLUMN tipo TO kind;
ALTER TABLE bill RENAME COLUMN iniciativa TO initiative;
ALTER TABLE bill RENAME COLUMN camara_origen TO origin_chamber;
ALTER TABLE bill RENAME COLUMN radicado_at TO filed_at;
ALTER TABLE bill RENAME COLUMN tema TO topic;

ALTER INDEX bill_numero_camara_uq RENAME TO bill_chamber_number_uq;
ALTER INDEX bill_numero_senado_uq RENAME TO bill_senate_number_uq;
ALTER INDEX bill_titulo_trgm RENAME TO bill_title_trgm;

ALTER TABLE bill_ponente RENAME TO bill_rapporteur;
ALTER TABLE bill_rapporteur RENAME COLUMN coordinador TO coordinator;
ALTER TRIGGER bill_ponente_no_update ON bill_rapporteur RENAME TO bill_rapporteur_no_update;

ALTER TABLE bill_document RENAME COLUMN gaceta_number TO gazette_number;

-- +goose StatementEnd

-- +goose down
-- +goose StatementBegin

ALTER TABLE bill_document RENAME COLUMN gazette_number TO gaceta_number;

ALTER TRIGGER bill_rapporteur_no_update ON bill_rapporteur RENAME TO bill_ponente_no_update;
ALTER TABLE bill_rapporteur RENAME COLUMN coordinator TO coordinador;
ALTER TABLE bill_rapporteur RENAME TO bill_ponente;

ALTER INDEX bill_title_trgm RENAME TO bill_titulo_trgm;
ALTER INDEX bill_senate_number_uq RENAME TO bill_numero_senado_uq;
ALTER INDEX bill_chamber_number_uq RENAME TO bill_numero_camara_uq;

ALTER TABLE bill RENAME COLUMN topic TO tema;
ALTER TABLE bill RENAME COLUMN filed_at TO radicado_at;
ALTER TABLE bill RENAME COLUMN origin_chamber TO camara_origen;
ALTER TABLE bill RENAME COLUMN initiative TO iniciativa;
ALTER TABLE bill RENAME COLUMN kind TO tipo;
ALTER TABLE bill RENAME COLUMN senate_number TO numero_senado;
ALTER TABLE bill RENAME COLUMN chamber_number TO numero_camara;
ALTER TABLE bill RENAME COLUMN title_norm TO titulo_norm;
ALTER TABLE bill RENAME COLUMN title TO titulo;

-- +goose StatementEnd
