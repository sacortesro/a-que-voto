-- =====================================================================
-- 11_seed_vocabularies.sql: Initial data for actor, stage and document_kind
-- Depends on: 01, 02
-- Re-runnable: existing rows are left untouched (ON CONFLICT DO NOTHING).
-- =====================================================================

-- +goose Up
-- +goose StatementBegin

BEGIN;

-- ---------------------------------------------------------------------
-- actor (must come first: every other insert is attributed to an actor)
-- ---------------------------------------------------------------------
INSERT INTO actor (kind, name) VALUES
  ('system', 'system:migration'),
  ('user',   'santiago')
ON CONFLICT (name) DO NOTHING;

SELECT set_config('app.actor_id',
  (SELECT id::text FROM actor WHERE name = 'system:migration'), true);

-- ---------------------------------------------------------------------
-- stage
-- Debates 1-4 apply to every bill. Debates 5-8 are the second round
-- (segunda vuelta) of actos legislativos only.
-- Terminal states that can happen at any point use sort_order 900+.
-- ---------------------------------------------------------------------
INSERT INTO stage (slug, label, sort_order) VALUES
  ('radicado',                 'Radicado',                                         10),
  ('publicado',                'Publicado en Gaceta del Congreso',                 20),
  ('asignado_comision',        'Asignado a comisión',                              30),
  ('primer_debate',            'Primer debate (comisión, cámara de origen)',       40),
  ('segundo_debate',           'Segundo debate (plenaria, cámara de origen)',      50),
  ('tercer_debate',            'Tercer debate (comisión, cámara revisora)',        60),
  ('cuarto_debate',            'Cuarto debate (plenaria, cámara revisora)',        70),
  ('quinto_debate',            'Quinto debate (segunda vuelta, acto legislativo)', 80),
  ('sexto_debate',             'Sexto debate (segunda vuelta, acto legislativo)',  90),
  ('septimo_debate',           'Séptimo debate (segunda vuelta, acto legislativo)',100),
  ('octavo_debate',            'Octavo debate (segunda vuelta, acto legislativo)', 110),
  ('conciliacion',             'Conciliación',                                     120),
  ('objeciones_presidenciales','Objeciones presidenciales',                        130),
  ('revision_corte',           'Revisión de la Corte Constitucional',              140),
  ('sancion_presidencial',     'Sanción presidencial',                             150),
  ('promulgado',               'Promulgado (Diario Oficial)',                      160),
  ('archivado',                'Archivado',                                        900),
  ('retirado',                 'Retirado por el autor',                            910)
ON CONFLICT (slug) DO NOTHING;

-- ---------------------------------------------------------------------
-- document_kind
-- ---------------------------------------------------------------------
INSERT INTO document_kind (slug, label) VALUES
  -- Filing
  ('texto_radicado',             'Texto radicado con exposición de motivos'),
  -- Ponencias, one per debate
  ('ponencia_primer_debate',     'Informe de ponencia para primer debate'),
  ('ponencia_segundo_debate',    'Informe de ponencia para segundo debate'),
  ('ponencia_tercer_debate',     'Informe de ponencia para tercer debate'),
  ('ponencia_cuarto_debate',     'Informe de ponencia para cuarto debate'),
  ('ponencia_quinto_debate',     'Informe de ponencia para quinto debate'),
  ('ponencia_sexto_debate',      'Informe de ponencia para sexto debate'),
  ('ponencia_septimo_debate',    'Informe de ponencia para séptimo debate'),
  ('ponencia_octavo_debate',     'Informe de ponencia para octavo debate'),
  -- Approved texts, one per debate
  ('texto_aprobado_primer_debate',  'Texto aprobado en primer debate'),
  ('texto_aprobado_segundo_debate', 'Texto aprobado en segundo debate'),
  ('texto_aprobado_tercer_debate',  'Texto aprobado en tercer debate'),
  ('texto_aprobado_cuarto_debate',  'Texto aprobado en cuarto debate'),
  ('texto_aprobado_quinto_debate',  'Texto aprobado en quinto debate'),
  ('texto_aprobado_sexto_debate',   'Texto aprobado en sexto debate'),
  ('texto_aprobado_septimo_debate', 'Texto aprobado en séptimo debate'),
  ('texto_aprobado_octavo_debate',  'Texto aprobado en octavo debate'),
  -- Conciliation and final steps
  ('informe_conciliacion',       'Informe de conciliación'),
  ('texto_conciliado',           'Texto conciliado'),
  ('objeciones_presidenciales',  'Escrito de objeciones presidenciales'),
  ('informe_objeciones',         'Informe sobre objeciones presidenciales'),
  ('sentencia_corte',            'Sentencia de la Corte Constitucional'),
  ('texto_definitivo',           'Texto definitivo (ley o acto legislativo)'),
  -- Supporting documents
  ('concepto',                   'Concepto (ministerio u otra entidad)'),
  ('acta',                       'Acta de sesión'),
  ('fe_de_erratas',              'Fe de erratas')
ON CONFLICT (slug) DO NOTHING;

COMMIT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Reverse order of Up: dependents first (document_kind, stage), actor last,
-- because every table references actor through created_by/updated_by.
BEGIN;

DELETE FROM document_kind
WHERE slug IN (
  'texto_radicado',
  'ponencia_primer_debate',
  'ponencia_segundo_debate',
  'ponencia_tercer_debate',
  'ponencia_cuarto_debate',
  'ponencia_quinto_debate',
  'ponencia_sexto_debate',
  'ponencia_septimo_debate',
  'ponencia_octavo_debate',
  'texto_aprobado_primer_debate',
  'texto_aprobado_segundo_debate',
  'texto_aprobado_tercer_debate',
  'texto_aprobado_cuarto_debate',
  'texto_aprobado_quinto_debate',
  'texto_aprobado_sexto_debate',
  'texto_aprobado_septimo_debate',
  'texto_aprobado_octavo_debate',
  'informe_conciliacion',
  'texto_conciliado',
  'objeciones_presidenciales',
  'informe_objeciones',
  'sentencia_corte',
  'texto_definitivo',
  'concepto',
  'acta',
  'fe_de_erratas'
);

DELETE FROM stage
WHERE slug IN (
  'radicado',
  'publicado',
  'asignado_comision',
  'primer_debate',
  'segundo_debate',
  'tercer_debate',
  'cuarto_debate',
  'quinto_debate',
  'sexto_debate',
  'septimo_debate',
  'octavo_debate',
  'conciliacion',
  'objeciones_presidenciales',
  'revision_corte',
  'sancion_presidencial',
  'promulgado',
  'archivado',
  'retirado'
);

DELETE FROM actor
WHERE name IN ('system:migration', 'santiago');

COMMIT;

-- +goose StatementEnd
