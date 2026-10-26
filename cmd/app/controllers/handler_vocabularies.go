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

func (apiCfg *APIConfig) HandleGetAllDocumentKinds(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	kindList, err := apiCfg.Store.Queries().GetDocumentKinds(ctx)
	if err != nil {
		utils.RespondWithInternalError(w, "get document kinds", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.DatabaseDocumentKindsToDocumentKinds(kindList))
}
