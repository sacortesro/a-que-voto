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

const (
	legislaturesPerTerm = 4
	termStartMonth      = time.July
	termStartDay        = 20
	termEndDay          = 19
	electionCycleYears  = 4
	electionCycleOffset = 2
)

func (apiCfg *APIConfig) HandleGetTerms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	terms, err := apiCfg.Store.Queries().GetTerms(ctx)
	if err != nil {
		utils.RespondWithInternalError(w, "get terms", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.DatabaseTermsToTerms(terms))
}

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

	startDate, err := utils.ParseDate(params.StartDate)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "start_date must be formatted as YYYY-MM-DD")
		return
	}
	if !isTermStart(startDate) {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("start_date must be %d July of an election year", termStartDay))
		return
	}

	endDate, err := utils.ParseDate(params.EndDate)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "end_date must be formatted as YYYY-MM-DD")
		return
	}
	if !isTermEnd(endDate) {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("end_date must be %d July", termEndDay))
		return
	}

	yearsInTerm := endDate.Year() - startDate.Year()
	if yearsInTerm != legislaturesPerTerm {
		utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("term must span %d years, got %d", legislaturesPerTerm, yearsInTerm))
		return
	}

	ctx := r.Context()

	var term database.Term
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		term, err = q.CreateTerm(ctx, database.CreateTermParams{
			Label:     periodLabel(startDate, endDate),
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err != nil {
			return fmt.Errorf("create term: %w", err)
		}

		for i := range legislaturesPerTerm {
			legislatureStart := startDate.AddDate(i, 0, 0)
			legislatureEnd := legislatureStart.AddDate(1, 0, 0).AddDate(0, 0, -1)

			_, err := q.CreateLegislature(ctx, database.CreateLegislatureParams{
				TermID:    term.ID,
				Label:     periodLabel(legislatureStart, legislatureEnd),
				StartDate: legislatureStart,
				EndDate:   legislatureEnd,
			})
			if err != nil {
				return fmt.Errorf("create legislature %d: %w", i+1, err)
			}
		}
		return nil
	})
	if err != nil {
		utils.RespondWithInternalError(w, "create term with legislatures", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, models.DatabaseTermToTerm(term))
}

func isTermStart(date time.Time) bool {
	return date.Month() == termStartMonth &&
		date.Day() == termStartDay &&
		(date.Year()+electionCycleOffset)%electionCycleYears == 0
}

func isTermEnd(date time.Time) bool {
	return date.Month() == termStartMonth && date.Day() == termEndDay
}

func periodLabel(start, end time.Time) string {
	return fmt.Sprintf("%d - %d", start.Year(), end.Year())
}
