package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/internal/database"
	"github.com/sacortesro/a-que-voto/internal/jobs"
	"github.com/sacortesro/a-que-voto/internal/store"
)

const (
	chamberRepresentatives = "1"
	chamberSenators        = "2"
	congressTermID         = "10"
	rowsPerPage            = "500"
	httpTimeout            = 30 * time.Second
	maxBodyBytes           = 20 << 20
)

type scrapePersons struct {
	name    string
	run     func(ctx context.Context) (jobs.Result, error)
	store   *store.Store
	chamber string
}

type scrapeData struct {
	BuildID string `json:"buildId"`
	Props   struct {
		PageProps personsProps `json:"pageProps"`
	} `json:"props"`
}

type personsProps struct {
	InitialList  []congressPerson `json:"ListaInicial"`
	InitialTotal int              `json:"TotalInicial"`
}

type congressPerson struct {
	PersonID   int32   `json:"persona_id"`
	FirstNames string  `json:"nombres"`
	LastNames  string  `json:"apellidos"`
	Photo      *string `json:"persona_imagen"`
	Party      *string `json:"partido"`
	PartyImage *string `json:"partido_imagen"`
	Committee  *string `json:"comision"`
	ChamberID  int     `json:"corporacion_id"`
	TermID     int     `json:"cuatrienio_id"`
	Active     int     `json:"activo"`
}

func (j scrapePersons) Name() string {
	switch j.chamber {
	case chamberRepresentatives:
		return "scrape-representatives"
	case chamberSenators:
		return "scrape-senadores"
	default:
		return "scrape-person"
	}
}

func (j scrapePersons) Run(ctx context.Context) (jobs.Result, error) {

	codeSource := "congresovisible"
	source, err := j.store.Queries().GetSourceByCode(ctx, codeSource)
	if err != nil {
		return jobs.Result{}, err
	}

	req, err := j.createRequest(ctx, source.BaseUrl+"/congresistas/", j.chamber)
	if err != nil {
		return jobs.Result{}, err
	}

	header, body, err := httpRequestPage(req)
	if err != nil {
		return jobs.Result{}, err
	}

	scrapeDocument, err := parseNextData[scrapeData](body)
	if err != nil {
		return jobs.Result{}, err
	}

	contentSha256, err := hashPersons(scrapeDocument.Props.PageProps.InitialList)
	if err != nil {
		return jobs.Result{}, err
	}

	changeCongressList, err := j.hasContentChanged(ctx, contentSha256, source, req)
	if err != nil {
		return jobs.Result{}, err
	}

	if !changeCongressList {
		return jobs.Result{ItemsSeen: len(scrapeDocument.Props.PageProps.InitialList), ItemsChanged: 0}, nil
	}

	runID, ok := jobs.RunID(ctx)
	if !ok {
		return jobs.Result{}, jobs.ErrNoRunID
	}

	_, err = j.createRawDocument(ctx, source, req, contentSha256, header, runID)

	if err != nil {
		return jobs.Result{}, err
	}

	personCount, err := j.createScrapedPersons(ctx, scrapeDocument)
	if err != nil {
		return jobs.Result{}, err
	}

	return jobs.Result{ItemsSeen: len(scrapeDocument.Props.PageProps.InitialList), ItemsChanged: personCount}, nil

}

func (j scrapePersons) hasContentChanged(ctx context.Context, contentHash []byte, source database.Source, req *http.Request) (bool, error) {

	prevHash, err := j.store.Queries().GetLatestRawDocumentHash(ctx, database.GetLatestRawDocumentHashParams{
		SourceID: source.ID,
		Url:      req.URL.String(),
	})

	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return !bytes.Equal(contentHash, prevHash), nil
}

func (j scrapePersons) createRawDocument(ctx context.Context,
	source database.Source,
	req *http.Request,
	contentSha256 []byte,
	header *http.Header,
	runID uuid.UUID,
) (database.RawDocument, error) {

	var rawDocument database.RawDocument
	err := j.store.Write(ctx, func(q *database.Queries) error {
		var err error
		rawDocument, err = q.CreateRawDocument(ctx, database.CreateRawDocumentParams{
			SourceID:      source.ID,
			Url:           req.URL.String(),
			ContentSha256: contentSha256,
			StorageKey:    "", // TODO
			ContentType:   sql.NullString{String: header.Get("Content-Type"), Valid: true},
			FetchedAt:     time.Now(),
			FetchRunID:    uuid.NullUUID{UUID: runID, Valid: true},
		})
		return err
	})
	if err != nil {
		return rawDocument, err
	}

	return rawDocument, nil
}

func (j scrapePersons) createScrapedPersons(ctx context.Context, scrapeDocument *scrapeData) (int, error) {

	n := len(scrapeDocument.Props.PageProps.InitialList)

	for _, person := range scrapeDocument.Props.PageProps.InitialList {
		fullName := person.FirstNames + " " + person.LastNames

		photo := sql.NullString{}
		if person.Photo != nil {
			photo = sql.NullString{String: *person.Photo, Valid: true}
		}

		err := j.store.Write(ctx, func(q *database.Queries) error {
			_, err := q.CreatePerson(ctx, database.CreatePersonParams{
				FullName:          fullName,
				BirthDate:         sql.NullTime{},
				PhotoUrl:          photo,
				SocialMedia:       json.RawMessage(`[]`),
				IDCongresoVisible: sql.NullInt32{Int32: person.PersonID, Valid: true},
			})
			return err
		})
		if err != nil {
			return 0, err
		}
	}

	return n, nil
}

func (scrapePersons) createRequest(ctx context.Context, baseUrl string, chamber string) (*http.Request, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("page", "1")
	q.Set("rows", rowsPerPage)
	q.Set("corporacion", chamber)
	q.Set("cuatrienio", congressTermID)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "AQV-scraper/0.1 aquevoto")

	return req, nil

}

// Hash of the persons list, used to detect changes in the source.
func hashPersons(list []congressPerson) ([]byte, error) {
	sorted := make([]congressPerson, len(list))
	copy(sorted, list)
	sort.Slice(sorted, func(a, b int) bool { return sorted[a].PersonID < sorted[b].PersonID })

	payload, err := json.Marshal(sorted)
	if err != nil {
		return nil, fmt.Errorf("marshal persons for hashing: %w", err)
	}

	h := sha256.New()
	h.Write(payload)
	return h.Sum(nil), nil
}

func httpRequestPage(req *http.Request) (*http.Header, []byte, error) {

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, nil, err
	}

	return &resp.Header, body, nil
}
