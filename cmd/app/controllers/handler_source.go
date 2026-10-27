package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/database"
)

func (apiCfg *APIConfig) HandleCreateSource(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Code     string `json:"code"`
		BaseURL  string `json:"base_url"`
		Priority int16  `json:"priority"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ctx := r.Context()

	var source database.Source
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		source, err = q.CreateSource(ctx, database.CreateSourceParams{
			Code:     params.Code,
			BaseUrl:  params.BaseURL,
			Priority: params.Priority,
		})
		return err
	})
	if err != nil {
		utils.RespondWithInternalError(w, "create source", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, models.DatabaseSourceToSource(source))
}

func (apiCfg *APIConfig) HandleGetAllSources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sourceList, err := apiCfg.Store.Queries().GetAllSources(ctx)

	if err != nil {
		utils.RespondWithInternalError(w, "get sources", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.DatabaseSourcesToSources(sourceList))
}
