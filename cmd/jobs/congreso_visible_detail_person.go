package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

type scrapeDetailData struct {
	Props struct {
		PageProps struct {
			Data personDetail `json:"data"`
		} `json:"pageProps"`
	} `json:"props"`
}

type personDetail struct {
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

	persons, err := j.store.Queries().GetPersonsCongresoVisible(ctx)
	if err != nil {
		return jobs.Result{}, err
	}

	updateCount, failed := 0, 0

	for i, person := range persons {

		// Pause between profiles so the site sees a slow reader, not a flood.
		if i > 0 {
			if err := waitBetweenRequests(ctx, requestDelay); err != nil {
				return jobs.Result{}, err
			}
		}

		req, err := j.createRequest(ctx, source.BaseUrl, "/congresistas/perfil/", person)
		if err != nil {
			return jobs.Result{}, err
		}

		_, body, err := httpRequestPage(req)
		if err != nil {
			return jobs.Result{}, err
		}

		scrapedData, err := parseNextData[scrapeDetailData](body)
		if err != nil {
			return jobs.Result{}, err
		}

		_, err = j.updatePersonDetail(ctx, scrapedData.Props.PageProps.Data, person.ID)
		if err != nil {
			log.Printf("update person %s: %v", person.ID, err)
			failed++
			continue
		}

		updateCount++
	}

	var partial string
	if failed > 0 {
		partial = fmt.Sprintf("%d of %d profiles failed to update", failed, len(persons))
	}

	return jobs.Result{ItemsSeen: len(persons), ItemsChanged: updateCount, Partial: partial}, nil
}

func (j scrapeDetailPerson) updatePersonDetail(ctx context.Context, detail personDetail, id uuid.UUID) (database.Person, error) {

	birth, err := time.Parse("2006-01-02", detail.BirthDate)
	if err != nil {
		return database.Person{}, err
	}

	socialMedia, err := json.Marshal(socialMediaToSocialLink(detail))
	if err != nil {
		return database.Person{}, err
	}

	var person database.Person
	err = j.store.Write(ctx, func(q *database.Queries) error {
		var err error
		person, err = q.UpdatePersonDetail(ctx, database.UpdatePersonDetailParams{
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

func (scrapeDetailPerson) createRequest(ctx context.Context, baseUrl string, complementUrl string, person database.GetPersonsCongresoVisibleRow) (*http.Request, error) {

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

func socialMediaToSocialLink(detail personDetail) []SocialLink {

	links := make([]SocialLink, 0, len(detail.SocialMedia))

	for _, s := range detail.SocialMedia {

		links = append(links, SocialLink{Platform: s.AccountData.Name, URL: s.Account})

	}

	return links

}
