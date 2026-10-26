package controllers

import (
	"net/http"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
)

func (apiCfg *APIConfig) HandleGetAllStages(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	stageList, err := apiCfg.Store.Queries().GetStages(ctx)
	if err != nil {
		utils.RespondWithInternalError(w, "get stages", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.DatabaseStagesToStages(stageList))
}
