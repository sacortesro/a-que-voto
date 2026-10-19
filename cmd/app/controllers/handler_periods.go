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

func (apiCfg *APIConfig) HandlerCreatePeriod(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, 400, fmt.Sprintf("Error parsing json %v", err))
		return
	}

	startDate, err := time.Parse("2006-01-02", params.StartDate)
	endDate, err := time.Parse("2006-01-02", params.EndDate)

	if startDate.Day() != 20 || startDate.Month() != time.July || (int(startDate.Year())+2)%4 > 0 {
		utils.RespondWithError(w, 400, fmt.Sprintf("Bad Date: %v", startDate))
		return
	}

	if startDate.Day() != 19 || startDate.Month() != time.July {
		utils.RespondWithError(w, 400, fmt.Sprintf("Bad Date: %v", endDate))
		return
	}

	if startDate.After(endDate) {
		utils.RespondWithError(w, 400, fmt.Sprintf("Bad Dates %v - %v", startDate, endDate))
	}

	durationTerm := int(endDate.AddDate(0, 0, 1).Sub(startDate).Hours() / 24 / 365)

	if durationTerm != 4 {
		utils.RespondWithError(w, 400, fmt.Sprintf("Time difference is not 4 years: %d", durationTerm))
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
		utils.RespondWithError(w, 500, fmt.Sprintf("Coudn't create term %v", err))
		return
	}

	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		for i := range durationTerm {

			legStart := startDate.AddDate(i, 0, 0)
			legEnd := legStart.AddDate(1, 0, 0).AddDate(0, 0, -1)
			label := fmt.Sprintf("%d - %d", legStart.Year(), legEnd.Year())

			_, err := q.CreateLegislatura(ctx, database.CreateLegislaturaParams{
				TermID:    term.ID,
				Label:     label,
				StartDate: legStart,
				EndDate:   legEnd,
			})
			if err != nil {
				return fmt.Errorf("create legislatura %q: %w", label, err)
			}
		}
		return nil
	})
	if err != nil {
		utils.RespondWithError(w, 500, fmt.Sprintf("Couldn't create legislaturas %v", err))
		return
	}

	utils.RespondWithJson(w, http.StatusCreated, models.DatabaseTermToTerm(term))
}
