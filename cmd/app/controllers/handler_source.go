package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/database"
)

func (apiCfg *APIConfig) HandleCreateSource(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Code     string `json:"code"`
		BaseUrl  string `json:"base_url"`
		Priority int16  `json:"priority"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		utils.RespondWithError(w, 400, fmt.Sprintf("Error parsing json %v", err))
		return
	}

	ctx := r.Context()

	var source database.Source
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		source, err = q.CreateSource(ctx, database.CreateSourceParams{
			Code:     params.Code,
			BaseUrl:  params.BaseUrl,
			Priority: params.Priority,
		})
		return err
	})
	if err != nil {
		utils.RespondWithError(w, 500, fmt.Sprintf("Couldn't create source %v", err))
		return
	}

	utils.RespondWithJson(w, http.StatusCreated, models.DatabaseSourceToSource(source))
}
