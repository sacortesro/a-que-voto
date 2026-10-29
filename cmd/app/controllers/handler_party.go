package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/database"
)

func (apiCfg *APIConfig) HandleGetParties(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	parties, err := apiCfg.Store.Queries().ListActiveParties(ctx)
	if err != nil {
		utils.RespondWithInternalError(w, "get parties", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.DatabasePartiesToParties(parties))
}

func (apiCfg *APIConfig) HandleCreateParty(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name    string  `json:"name"`
		Acronym *string `json:"acronym"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ctx := r.Context()

	var party database.Party
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		party, err = q.CreateParty(ctx, database.CreatePartyParams{
			Name:    params.Name,
			Acronym: utils.NullString(params.Acronym),
		})
		return err
	})
	if err != nil {
		utils.RespondWithInternalError(w, "create party", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, models.DatabasePartyToParty(party))

}

func (apiCfg *APIConfig) HandleUpdateParty(w http.ResponseWriter, r *http.Request) {

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid party id")
		return
	}

	type parameters struct {
		Name     string  `json:"name"`
		Acronym  *string `json:"acronym"`
		IsActive *bool   `json:"is_active"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if params.IsActive == nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid status")
		return
	}
	if strings.TrimSpace(params.Name) == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid name")
		return
	}

	ctx := r.Context()
	var party database.Party
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		party, err = q.UpdateParty(ctx, database.UpdatePartyParams{
			ID:       id,
			Name:     strings.TrimSpace(params.Name),
			Acronym:  utils.NullString(params.Acronym),
			IsActive: *params.IsActive,
		})
		return err
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, http.StatusNotFound, "party not found")
			return
		}
		if utils.IsUniqueViolation(err) {
			utils.RespondWithError(w, http.StatusConflict, "Name already exists")
			return
		}

		utils.RespondWithInternalError(w, "update party", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.DatabasePartyToParty(party))
}
