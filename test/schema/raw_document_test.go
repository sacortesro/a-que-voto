package schema_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
	"time"

	"github.com/sacortesro/a-que-voto/internal/database"
	"github.com/sacortesro/a-que-voto/test/dbtest"
)

func TestRawDocumentHasNoStorageKey(t *testing.T) {
	db := dbtest.Open(t)

	var n int
	err := db.QueryRow(`SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'raw_document' AND column_name = 'storage_key'`).Scan(&n)
	if err != nil {
		t.Fatalf("query columns: %v", err)
	}
	if n != 0 {
		t.Fatalf("raw_document.storage_key columns = %d, want 0", n)
	}

	// Control: the same query finds a column that exists, so 0 above is real.
	err = db.QueryRow(`SELECT count(*) FROM information_schema.columns
		WHERE table_name = 'raw_document' AND column_name = 'content_type'`).Scan(&n)
	if err != nil || n != 1 {
		t.Fatalf("control column count = %d, err = %v, want 1", n, err)
	}
}

func TestCreateRawDocumentWithoutStorageKey(t *testing.T) {
	db := dbtest.Open(t)
	ctx := context.Background()

	tests := map[string]sql.NullString{
		"with content type":    {String: "text/html", Valid: true},
		"without content type": {},
	}
	for name, contentType := range tests {
		t.Run(name, func(t *testing.T) {
			tx := dbtest.Begin(t, db)
			q := database.New(tx)

			var sourceID int16
			err := tx.QueryRow(`INSERT INTO source (code, base_url, priority) VALUES ('raw-doc', 'https://example.test', 1) RETURNING id`).Scan(&sourceID)
			if err != nil {
				t.Fatalf("create source: %v", err)
			}

			sum := sha256.Sum256([]byte("content"))
			doc, err := q.CreateRawDocument(ctx, database.CreateRawDocumentParams{
				SourceID:      sourceID,
				Url:           "https://example.test/doc",
				ContentSha256: sum[:],
				ContentType:   contentType,
				FetchedAt:     time.Now(),
			})
			if err != nil {
				t.Fatalf("create raw document: %v", err)
			}
			if doc.Url != "https://example.test/doc" || doc.ContentType != contentType {
				t.Fatalf("document = %+v, want url and content type %+v", doc, contentType)
			}
		})
	}
}
