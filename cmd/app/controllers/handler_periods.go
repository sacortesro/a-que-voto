package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/database"
)

const legislaturesPerTerm = 4

func (apiCfg *APIConfig) HandleCreatePeriod(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	startDate, err := time.Parse("2006-01-02", params.StartDate)
	if startDate.Day() != 20 || startDate.Month() != time.July || (int(startDate.Year())+2)%4 > 0 {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Bad Date: %v", startDate))
		return
	}

	endDate, err := time.Parse("2006-01-02", params.EndDate)
	if endDate.Day() != 19 || endDate.Month() != time.July {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Bad Date: %v", endDate))
		return
	}

	if startDate.After(endDate) {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Bad Dates %v - %v", startDate, endDate))
		return
	}

	durationTerm := int(endDate.AddDate(0, 0, 1).Sub(startDate).Hours() / 24 / 365)

	if durationTerm != legislaturesPerTerm {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Time difference is not %d years: %d", legislaturesPerTerm, durationTerm))
		return
	}

	label := fmt.Sprintf("%d - %d", startDate.Year(), endDate.Year())

	ctx := r.Context()

	var term database.Term
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		term, err = q.CreateTerm(ctx, database.CreateTermParams{
			Label:     label,
			StartDate: startDate,
			EndDate:   endDate,
		})
		return err
	})

	if err != nil {
		utils.RespondWithInternalError(w, "create term", err)
		return
	}

	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		for i := range durationTerm {

			legStart := startDate.AddDate(i, 0, 0)
			legEnd := legStart.AddDate(1, 0, 0).AddDate(0, 0, -1)
			label := fmt.Sprintf("%d - %d", legStart.Year(), legEnd.Year())

			_, err := q.CreateLegislature(ctx, database.CreateLegislatureParams{
				TermID:    term.ID,
				Label:     label,
				StartDate: legStart,
				EndDate:   legEnd,
			})
			if err != nil {
				return fmt.Errorf("create legislature %q: %w", label, err)
			}
		}
		return nil
	})
	if err != nil {
		utils.RespondWithInternalError(w, "create legislatures", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, models.DatabaseTermToTerm(term))
}
