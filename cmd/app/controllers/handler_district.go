package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/database"
)

func (apiCfg *APIConfig) HandleListDistricts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	districts, err := apiCfg.Store.Queries().ListDistricts(ctx)
	if err != nil {
		utils.RespondWithInternalError(w, "list districts", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.DatabaseDistrictsToDistricts(districts))
}

func (apiCfg *APIConfig) HandleCreateDistrict(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Chamber    string  `json:"chamber"`
		Kind       string  `json:"kind"`
		Department *string `json:"department"`
		Name       string  `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ctx := r.Context()

	var district database.District
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		district, err = q.CreateDistrict(ctx, database.CreateDistrictParams{
			Chamber:    params.Chamber,
			Kind:       params.Kind,
			Department: utils.NullString(params.Department),
			Name:       params.Name,
		})
		return err
	})
	if err != nil {
		utils.RespondWithInternalError(w, "create district", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, models.DatabaseDistrictToDistrict(district))
}
