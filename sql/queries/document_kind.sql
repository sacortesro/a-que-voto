-- name: GetDocumentKinds :many
SELECT *
FROM document_kind
ORDER BY slug ASC;
