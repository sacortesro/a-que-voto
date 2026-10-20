// Package jobs runs scraper jobs. A job only does its work; the runner owns
// the job's identity (actor), its fetch_run record and its error handling.
// Nothing here depends on the scheduler, so GitHub Actions can be swapped
// for another trigger without touching the jobs.
package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/internal/actors"
	"github.com/sacortesro/a-que-voto/internal/auth"
	"github.com/sacortesro/a-que-voto/internal/database"
	"github.com/sacortesro/a-que-voto/internal/store"
)

// staleAfter is how old a `running` fetch_run must be before a new run of
// the same job marks it failed. It matches GitHub Actions' default job
// timeout, so a live run is never taken for a dead one.
const staleAfter = 6 * time.Hour

const (
	StatusRunning = "running"
	StatusOK      = "ok"
	StatusFailed  = "failed"
	StatusPartial = "partial"
)

var (
	ErrUnknownJob   = errors.New("unknown job")
	ErrDuplicateJob = errors.New("duplicate job name")
	ErrNoRunID      = errors.New("no run id in context")
)

// Result is what a job reports back.
// A non-empty Partial means the run finished with some errors:
// status becomes `partial` and the message is stored in fetch_run.error.
type Result struct {
	ItemsSeen    int
	ItemsChanged int
	Partial      string
}

type Job interface {
	Name() string
	Run(ctx context.Context) (Result, error)
}

// Registry holds the jobs the runner may execute. Unregistered names are
// refused, so a typo can never create an actor row.
type Registry struct {
	jobs map[string]Job
}

func NewRegistry(list ...Job) (*Registry, error) {
	r := &Registry{jobs: make(map[string]Job, len(list))}
	for _, j := range list {
		if _, exists := r.jobs[j.Name()]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateJob, j.Name())
		}
		r.jobs[j.Name()] = j
	}
	return r, nil
}

func (r *Registry) Get(name string) (Job, bool) {
	j, ok := r.jobs[name]
	return j, ok
}

type Runner struct {
	store    *store.Store
	registry *Registry
	now      func() time.Time
}

func NewRunner(s *store.Store, r *Registry) *Runner {
	return &Runner{store: s, registry: r, now: time.Now}
}

// Run executes one job by name. The actor is derived from the name
// ("job:<name>"), so every write the job makes is attributed to it
// without the job knowing.
func (r *Runner) Run(ctx context.Context, name string) error {
	job, ok := r.registry.Get(name)
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownJob, name)
	}

	actor, err := actors.Ensure(ctx, r.store.Queries(), actors.KindJob, "job:"+name)
	if err != nil {
		return err
	}
	ctx = auth.WithActor(ctx, actor)

	if err := r.store.Write(ctx, func(q *database.Queries) error {
		_, err := q.FailStaleRuns(ctx, database.FailStaleRunsParams{
			Job:           name,
			StartedBefore: r.now().Add(-staleAfter),
		})
		return err
	}); err != nil {
		return fmt.Errorf("fail stale runs: %w", err)
	}

	var run database.FetchRun
	if err := r.store.Write(ctx, func(q *database.Queries) error {
		var err error
		run, err = q.CreateRun(ctx, database.CreateRunParams{
			Job:       name,
			StartedAt: r.now(),
			Status:    StatusRunning,
		})
		return err
	}); err != nil {
		return fmt.Errorf("open run: %w", err)
	}

	ctx = context.WithValue(ctx, runIDKey{}, run.ID)

	result, runErr := runSafely(ctx, job)
	status, msg := outcome(result, runErr)

	// The job's ctx may already be cancelled (timeout, SIGTERM). The run
	// must still be closed, or it stays `running` until the next stale sweep.
	closeCtx := context.WithoutCancel(ctx)
	if err := r.store.Write(closeCtx, func(q *database.Queries) error {
		_, err := q.UpdateRun(closeCtx, database.UpdateRunParams{
			ID:           run.ID,
			FinishedAt:   sql.NullTime{Time: r.now(), Valid: true},
			Status:       status,
			ItemsSeen:    sql.NullInt32{Int32: int32(result.ItemsSeen), Valid: true},
			ItemsChanged: sql.NullInt32{Int32: int32(result.ItemsChanged), Valid: true},
			Error:        sql.NullString{String: msg, Valid: msg != ""},
		})
		return err
	}); err != nil {
		return errors.Join(runErr, fmt.Errorf("close run: %w", err))
	}

	return runErr
}

// runSafely turns a panic into an ordinary error, so the run is still closed
// as `failed` instead of staying `running`.
func runSafely(ctx context.Context, job Job) (result Result, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("job panicked: %v", p)
		}
	}()
	return job.Run(ctx)
}

func outcome(res Result, runErr error) (status, msg string) {
	switch {
	case runErr != nil:
		return StatusFailed, runErr.Error()
	case res.Partial != "":
		return StatusPartial, res.Partial
	default:
		return StatusOK, ""
	}
}

type runIDKey struct{}

func RunID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(runIDKey{}).(uuid.UUID)
	return id, ok
}
