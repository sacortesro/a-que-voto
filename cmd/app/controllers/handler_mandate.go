package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/database"
)

func (apiCfg *APIConfig) HandlerCreateMandate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Person             string  `json:"person"`
		Term               string  `json:"term"`
		Chamber            string  `json:"chamber"`
		DistrictKind       string  `json:"district_kind"`
		DistrictDepartment *string `json:"district_department"`
		Party              string  `json:"party"`
		StartDate          string  `json:"start_date"`
		EndDate            *string `json:"end_date"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if params.Chamber != ChamberSenate && params.Chamber != ChamberRepresentative {
		utils.RespondWithError(w, http.StatusBadRequest, "chamber not found")
		return
	}

	startDate, err := utils.ParseDate(params.StartDate)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "start_date must be formatted as YYYY-MM-DD")
		return
	}

	endDate, err := utils.ParseOptionalDate(params.EndDate)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "end_date must be formatted as YYYY-MM-DD")
		return
	}
	if endDate != nil && endDate.Before(startDate) {
		utils.RespondWithError(w, http.StatusBadRequest, "end_date must not be before start_date")
		return
	}

	ctx := r.Context()
	queries := apiCfg.Store.Queries()

	personID, err := queries.GetPersonIDByName(ctx, params.Person)
	if err != nil {
		respondLookupError(w, err, "person not found", "lookup person")
		return
	}

	termID, err := queries.GetTermByLabel(ctx, params.Term)
	if err != nil {
		respondLookupError(w, err, "term not found", "lookup term")
		return
	}

	districtID, err := queries.GetDistrict(ctx, database.GetDistrictParams{
		Chamber:    params.Chamber,
		Kind:       params.DistrictKind,
		Department: utils.NullString(params.DistrictDepartment),
	})
	if err != nil {
		respondLookupError(w, err, "district not found", "lookup district")
		return
	}

	partyID, err := queries.GetPartyByName(ctx, params.Party)
	if err != nil {
		respondLookupError(w, err, "party not found", "lookup party")
		return
	}

	var mandate database.Mandate
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		mandate, err = q.CreatePersonMandate(ctx, database.CreatePersonMandateParams{
			PersonID:   personID,
			TermID:     termID,
			Chamber:    params.Chamber,
			DistrictID: utils.NullInt16(&districtID),
			PartyID:    utils.NullUUID(&partyID),
			StartDate:  startDate,
			EndDate:    utils.NullTime(endDate),
			Status:     StatusActive,
		})
		return err
	})
	if err != nil {
		utils.RespondWithInternalError(w, "create mandate", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, models.DatabaseMandateToMandate(mandate))
}

// respondLookupError answers 400 when the referenced row does not exist,
// and 500 for any other failure, so a broken database is not reported as bad input.
func respondLookupError(w http.ResponseWriter, err error, notFoundMsg, op string) {
	if errors.Is(err, sql.ErrNoRows) {
		utils.RespondWithError(w, http.StatusBadRequest, notFoundMsg)
		return
	}
	utils.RespondWithInternalError(w, op, err)
}
