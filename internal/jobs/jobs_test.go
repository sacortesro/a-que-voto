package jobs

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/internal/actors"
	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/internal/database"
	"github.com/sacortesro/a-que-voto/internal/store"

	_ "github.com/lib/pq"
)

// These tests need a migrated database. Set AQV_TEST_DB_URL to run them;
// otherwise they are skipped. Each test uses its own job name, so runs
// do not interfere with each other.
func testStore(t *testing.T) (*store.Store, *sql.DB) {
	t.Helper()
	dsn := os.Getenv("AQV_TEST_DB_URL")
	if dsn == "" {
		t.Skip("AQV_TEST_DB_URL not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return store.New(db), db
}

type fakeJob struct {
	name string
	run  func(ctx context.Context) (Result, error)
}

func (f fakeJob) Name() string                            { return f.name }
func (f fakeJob) Run(ctx context.Context) (Result, error) { return f.run(ctx) }

func newName(t *testing.T) string {
	return "test." + strings.ReplaceAll(t.Name(), "/", ".") + "." + uuid.NewString()[:8]
}

type runRow struct {
	status    string
	errText   sql.NullString
	createdBy uuid.UUID
}

func latestRun(t *testing.T, db *sql.DB, job string) runRow {
	t.Helper()
	var r runRow
	err := db.QueryRow(
		`SELECT status, error, created_by FROM fetch_run
		 WHERE job = $1 ORDER BY started_at DESC LIMIT 1`, job,
	).Scan(&r.status, &r.errText, &r.createdBy)
	if err != nil {
		t.Fatalf("read fetch_run for %s: %v", job, err)
	}
	return r
}

func actorID(t *testing.T, db *sql.DB, name string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := db.QueryRow(`SELECT id FROM actor WHERE name = $1`, "job:"+name).Scan(&id); err != nil {
		t.Fatalf("read actor job:%s: %v", name, err)
	}
	return id
}

func TestRunAttributesWritesToJobActor(t *testing.T) {
	s, db := testStore(t)
	name := newName(t)
	code := "src-" + uuid.NewString()

	var seenActor auth.Actor
	job := fakeJob{name: name, run: func(ctx context.Context) (Result, error) {
		seenActor, _ = auth.ActorFromContext(ctx)
		err := s.Write(ctx, func(q *database.Queries) error {
			_, err := q.CreateSource(ctx, database.CreateSourceParams{
				Code: code, BaseUrl: "https://example.test", Priority: 1,
			})
			return err
		})
		return Result{ItemsSeen: 1, ItemsChanged: 1}, err
	}}

	reg, err := NewRegistry(job)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewRunner(s, reg).Run(context.Background(), name); err != nil {
		t.Fatalf("Run: %v", err)
	}

	want := actorID(t, db, name)
	if seenActor.ID != want || seenActor.Kind != actors.KindJob {
		t.Fatalf("job ctx actor = %+v, want job actor %s", seenActor, want)
	}

	var sourceBy uuid.UUID
	if err := db.QueryRow(`SELECT created_by FROM source WHERE code = $1`, code).Scan(&sourceBy); err != nil {
		t.Fatalf("read source: %v", err)
	}
	if sourceBy != want {
		t.Errorf("source.created_by = %s, want %s", sourceBy, want)
	}

	run := latestRun(t, db, name)
	if run.status != "ok" || run.createdBy != want {
		t.Errorf("fetch_run = %+v, want status ok by %s", run, want)
	}
}

func TestRunFailureIsRecordedEvenWhenCtxCancelled(t *testing.T) {
	s, db := testStore(t)
	name := newName(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	boom := errors.New("boom")
	job := fakeJob{name: name, run: func(context.Context) (Result, error) {
		cancel() // the job's context dies mid-run; the run must still be closed
		return Result{}, boom
	}}

	reg, _ := NewRegistry(job)
	err := NewRunner(s, reg).Run(ctx, name)
	if !errors.Is(err, boom) {
		t.Fatalf("Run error = %v, want %v", err, boom)
	}

	run := latestRun(t, db, name)
	if run.status != "failed" || !run.errText.Valid || run.errText.String != "boom" {
		t.Errorf("fetch_run = %+v, want failed with error boom", run)
	}
}

func TestRunPanicBecomesFailedRun(t *testing.T) {
	s, db := testStore(t)
	name := newName(t)

	job := fakeJob{name: name, run: func(context.Context) (Result, error) {
		panic("kaboom")
	}}

	reg, _ := NewRegistry(job)
	if err := NewRunner(s, reg).Run(context.Background(), name); err == nil {
		t.Fatal("Run returned nil error for a panicking job")
	}

	run := latestRun(t, db, name)
	if run.status != "failed" || !strings.Contains(run.errText.String, "kaboom") {
		t.Errorf("fetch_run = %+v, want failed mentioning kaboom", run)
	}
}

func TestRunPartialIsNotAnError(t *testing.T) {
	s, db := testStore(t)
	name := newName(t)

	job := fakeJob{name: name, run: func(context.Context) (Result, error) {
		return Result{ItemsSeen: 3, ItemsChanged: 2, Partial: "2 pages failed"}, nil
	}}

	reg, _ := NewRegistry(job)
	if err := NewRunner(s, reg).Run(context.Background(), name); err != nil {
		t.Fatalf("Run: %v", err)
	}

	run := latestRun(t, db, name)
	if run.status != "partial" || run.errText.String != "2 pages failed" {
		t.Errorf("fetch_run = %+v, want partial with message", run)
	}
}

func TestRunRefusesUnregisteredJobWithoutCreatingActor(t *testing.T) {
	s, db := testStore(t)
	name := newName(t)

	reg, _ := NewRegistry()
	err := NewRunner(s, reg).Run(context.Background(), name)
	if !errors.Is(err, ErrUnknownJob) {
		t.Fatalf("Run error = %v, want ErrUnknownJob", err)
	}

	var count int
	if err := db.QueryRow(`SELECT count(*) FROM actor WHERE name = $1`, "job:"+name).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("actor rows for unknown job = %d, want 0", count)
	}
}

func TestRunMarksStaleRunningRowsFailed(t *testing.T) {
	s, db := testStore(t)
	name := newName(t)

	// A previous process died mid-run and left its row `running`.
	ctx := context.Background()
	jobActor, err := actors.Ensure(ctx, s.Queries(), actors.KindJob, "job:"+name)
	if err != nil {
		t.Fatal(err)
	}
	stale := auth.WithActor(ctx, jobActor)
	if err := s.Write(stale, func(q *database.Queries) error {
		_, err := q.CreateRun(stale, database.CreateRunParams{
			Job:       name,
			StartedAt: time.Now().Add(-7 * time.Hour),
			Status:    "running",
		})
		return err
	}); err != nil {
		t.Fatal(err)
	}

	job := fakeJob{name: name, run: func(context.Context) (Result, error) {
		return Result{}, nil
	}}
	reg, _ := NewRegistry(job)
	if err := NewRunner(s, reg).Run(ctx, name); err != nil {
		t.Fatalf("Run: %v", err)
	}

	var status string
	var errText sql.NullString
	if err := db.QueryRow(
		`SELECT status, error FROM fetch_run
		 WHERE job = $1 AND started_at < now() - interval '6 hours'`, name,
	).Scan(&status, &errText); err != nil {
		t.Fatalf("read stale run: %v", err)
	}
	if status != "failed" || !strings.Contains(errText.String, "stale") {
		t.Errorf("stale run = %s / %q, want failed / stale", status, errText.String)
	}
}
