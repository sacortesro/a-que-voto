package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/internal/database"
	"github.com/sacortesro/a-que-voto/internal/jobs"
	"github.com/sacortesro/a-que-voto/internal/store"
)

type scrapeDetailPerson struct {
	name  string
	run   func(ctx context.Context) (jobs.Result, error)
	store *store.Store
}

type scrapDetailData struct {
	Props struct {
		PageProps struct {
			Data detailCongresista `json:"data"`
		} `json:"pageProps"`
	} `json:"props"`
}

type detailCongresista struct {
	BirthDate   string `json:"fechaNacimiento"`
	SocialMedia []struct {
		Account     string `json:"cuenta"`
		AccountData struct {
			Name string `json:"nombre"`
		} `json:"datos_contacto"`
	} `json:"contactos"`
}

type SocialLink struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

// requestDelay is the minimum pause between two profile requests.
const requestDelay = 2 * time.Second

func (j scrapeDetailPerson) Name() string { return "scrape-detail-person" }

// waitBetweenRequests blocks for d, or returns early if ctx is cancelled.
func waitBetweenRequests(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (j scrapeDetailPerson) Run(ctx context.Context) (jobs.Result, error) {

	codeSource := "congresovisible"
	source, err := j.store.Queries().GetSourceByCode(ctx, codeSource)
	if err != nil {
		return jobs.Result{}, err
	}

	persons, err := j.store.Queries().GetPersons(ctx)
	if err != nil {
		return jobs.Result{}, err
	}

	// coment after debugging.
	// persons = persons[0:1]

	updateCount := 0

	for i, person := range persons {

		// Pause between profiles so the site sees a slow reader, not a flood.
		if i > 0 {
			if err := waitBetweenRequests(ctx, requestDelay); err != nil {
				return jobs.Result{}, err
			}
		}

		req, err := j.createRequest(ctx, source.BaseUrl, "/congresistas/perfil/", person) // Request Senadores
		if err != nil {
			return jobs.Result{}, err
		}

		_, body, err := httpRequestPage(req)
		if err != nil {
			return jobs.Result{}, err
		}

		var scrapDataRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__"[^>]*>(.*?)</script>`)

		scrapedData, err := j.parseScrapData(body, scrapDataRe)
		if err != nil {
			return jobs.Result{}, err
		}

		_, err = j.updateDetailCongresista(ctx, scrapedData.Props.PageProps.Data, person.ID)
		if err != nil {
			continue
		}

		fmt.Printf("updating...\n")
		fmt.Printf("record: %d \n", i)
		updateCount++
	}

	var partial string
	if updateCount != len(persons) {
		partial = "Partial update of records"
	}

	return jobs.Result{ItemsSeen: len(persons), ItemsChanged: updateCount, Partial: partial}, nil
}

func (j scrapeDetailPerson) updateDetailCongresista(ctx context.Context, detCongresista detailCongresista, id uuid.UUID) (database.Person, error) {

	birth, err := time.Parse("2006-01-02", detCongresista.BirthDate)
	if err != nil {
		return database.Person{}, err
	}

	socialMedia, err := json.Marshal(socialMediaToSocialLink(detCongresista))
	if err != nil {
		return database.Person{}, err
	}

	var person database.Person
	err = j.store.Write(ctx, func(q *database.Queries) error {
		var err error
		person, err = q.UpdatePerson(ctx, database.UpdatePersonParams{
			BirthDate:   sql.NullTime{Time: birth, Valid: true},
			SocialMedia: socialMedia,
			ID:          id,
		})
		return err
	})

	if err != nil {
		return person, err
	}
	return person, nil
}

func (scrapeDetailPerson) createRequest(ctx context.Context, baseUrl string, complementUrl string, person database.GetPersonsRow) (*http.Request, error) {

	nameSlug := strings.ReplaceAll(person.NameNorm.String, " ", "-")
	idStr := strconv.Itoa(int(person.IDCongresoVisible.Int32))

	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	u = u.JoinPath(complementUrl, nameSlug, idStr)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "AQV-scraper/0.1 aquevoto")

	return req, nil

}

func (scrapeDetailPerson) parseScrapData(html []byte, regex *regexp.Regexp) (*scrapDetailData, error) {
	m := regex.FindSubmatch(html)
	if m == nil {
		return nil, errors.New("__NEXT_DATA__ not found")
	}
	var nd scrapDetailData
	if err := json.Unmarshal(m[1], &nd); err != nil {
		return nil, fmt.Errorf("decode __NEXT_DATA__: %w", err)
	}
	return &nd, nil
}

func socialMediaToSocialLink(detContresista detailCongresista) []SocialLink {

	links := make([]SocialLink, 0, len(detContresista.SocialMedia))

	for _, s := range detContresista.SocialMedia {

		links = append(links, SocialLink{Platform: s.AccountData.Name, URL: s.Account})

	}

	return links

}
