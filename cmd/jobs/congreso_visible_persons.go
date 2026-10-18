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
	"regexp"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/internal/database"
	"github.com/sacortesro/a-que-voto/internal/jobs"
	"github.com/sacortesro/a-que-voto/internal/store"
)

type scrapePersons struct {
	name  string
	run   func(ctx context.Context) (jobs.Result, error)
	store *store.Store
}

type scrapData struct {
	BuildID string `json:"buildId"`
	Props   struct {
		PageProps congresistasProps `json:"pageProps"`
	} `json:"props"`
}

type congresistasProps struct {
	ListaInicial []CVCongresista `json:"ListaInicial"`
	TotalInicial int             `json:"TotalInicial"`
}

type CVCongresista struct {
	PersonaID     int32   `json:"persona_id"`
	Nombres       string  `json:"nombres"`
	Apellidos     string  `json:"apellidos"`
	PersonaImagen *string `json:"persona_imagen"`
	Partido       *string `json:"partido"`
	PartidoImagen *string `json:"partido_imagen"`
	Comision      *string `json:"comision"`
	CorporacionID int     `json:"corporacion_id"`
	CuatrienioID  int     `json:"cuatrienio_id"`
	Activo        int     `json:"activo"`
}

func (scrapePersons) Name() string { return "scrape-candidates" }

func (j scrapePersons) Run(ctx context.Context) (jobs.Result, error) {

	q := j.store.Queries()

	codeSource := "congresovisible"
	source, err := q.GetSourceByCode(ctx, codeSource)
	if err != nil {
		return jobs.Result{}, err
	}

	req, err := createRequest(ctx, source.BaseUrl+"/congresistas/")
	if err != nil {
		return jobs.Result{}, err
	}

	fmt.Printf("CreateRequest: %v", req)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return jobs.Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return jobs.Result{}, fmt.Errorf("unexpected status %s", resp.Status)
	}

	fmt.Printf("\n Status code request: %v", resp.StatusCode)

	body, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return jobs.Result{}, err
	}

	runId, ok := jobs.RunID(ctx)

	if !ok {
		return jobs.Result{}, errors.New("Fetch run id missing")
	}

	var scrapDataRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__"[^>]*>(.*?)</script>`)

	scrapDocument, err := parseScrapData(body, scrapDataRe)
	if err != nil {
		return jobs.Result{}, err
	}

	contentSha256, err := hashCongresistas(scrapDocument.Props.PageProps.ListaInicial)
	if err != nil {
		return jobs.Result{}, err
	}

	changeCongressList, err := j.hasContentChanged(ctx, contentSha256, source)

	if err != nil {
		return jobs.Result{}, err
	}

	if !changeCongressList {
		return jobs.Result{ItemsSeen: len(scrapDocument.Props.PageProps.ListaInicial), ItemsChanged: 0}, nil
	}

	_, err = j.createRawDocument(ctx, source, req, contentSha256, resp, runId)

	if err != nil {
		return jobs.Result{}, err
	}

	personCount, err := j.createScrapedPersons(ctx, scrapDocument)
	if err != nil {
		return jobs.Result{}, err
	}

	return jobs.Result{ItemsSeen: len(scrapDocument.Props.PageProps.ListaInicial), ItemsChanged: personCount}, nil

}

func (j scrapePersons) hasContentChanged(ctx context.Context, contentHash []byte, source database.Source) (bool, error) {

	prevHash, err := j.store.Queries().GetRawDocumentById(ctx, source.ID)

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
	resp *http.Response,
	runId uuid.UUID) (database.RawDocument, error) {

	var rawDocument database.RawDocument
	err := j.store.Write(ctx, func(q *database.Queries) error {
		var err error
		rawDocument, err = q.CreateRawDocument(ctx, database.CreateRawDocumentParams{
			SourceID:      source.ID,
			Url:           req.URL.String(),
			ContentSha256: contentSha256,
			StorageKey:    "", // TODO
			ContentType:   sql.NullString{String: resp.Header.Get("Content-Type"), Valid: true},
			FetchedAt:     time.Now(),
			FetchRunID:    uuid.NullUUID{UUID: runId, Valid: true},
		})
		return err
	})
	if err != nil {
		return rawDocument, err
	}

	return rawDocument, nil
}

func (j scrapePersons) createScrapedPersons(ctx context.Context, scrapDocument *scrapData) (int, error) {

	n := len(scrapDocument.Props.PageProps.ListaInicial)

	for i := 0; i < n; i++ {
		iPerson := scrapDocument.Props.PageProps.ListaInicial[i]

		var person database.Person
		err := j.store.Write(ctx, func(q *database.Queries) error {
			var err error
			fullName := iPerson.Nombres + " " + iPerson.Apellidos

			photo := sql.NullString{}
			if iPerson.PersonaImagen != nil {
				photo = sql.NullString{String: *iPerson.PersonaImagen, Valid: true}
			}

			person, err = q.CreatePerson(ctx, database.CreatePersonParams{
				FullName:          fullName,
				BirthDate:         sql.NullTime{},
				PhotoUrl:          photo,
				SocialMedia:       json.RawMessage(`[]`),
				IDCongresoVisible: sql.NullInt32{Int32: iPerson.PersonaID, Valid: true},
			})
			return err
		})

		if err != nil {
			return 0, err
		}

		_ = person
	}

	return n, nil
}

func createRequest(ctx context.Context, baseUrl string) (*http.Request, error) {
	u, _ := url.Parse(baseUrl)
	q := url.Values{}
	q.Set("page", "1")
	q.Set("rows", "500")
	q.Set("corporacion", "2") // Corporacion 2: Senadores, Corporación 1: Camara de representantes
	q.Set("cuartenio", "10")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "AQV-scraper/0.1 aquevoto")

	return req, nil

}

func parseScrapData(html []byte, regex *regexp.Regexp) (*scrapData, error) {
	m := regex.FindSubmatch(html)
	if m == nil {
		return nil, errors.New("__NEXT_DATA__ not found")
	}
	var nd scrapData
	if err := json.Unmarshal(m[1], &nd); err != nil {
		return nil, fmt.Errorf("decode __NEXT_DATA__: %w", err)
	}
	return &nd, nil
}

// Hash created to compare if there is a change in the congresistas data.
func hashCongresistas(list []CVCongresista) ([]byte, error) {
	sorted := make([]CVCongresista, len(list))
	copy(sorted, list)
	sort.Slice(sorted, func(a, b int) bool { return sorted[a].PersonaID < sorted[b].PersonaID })

	payload, err := json.Marshal(sorted)
	if err != nil {
		return nil, fmt.Errorf("marshal congresistas for hashing: %w", err)
	}

	h := sha256.New()
	h.Write(payload)
	return h.Sum(nil), nil
}
