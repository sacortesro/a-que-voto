package controllers

import (
	"encoding/json"
	"net/http"

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
