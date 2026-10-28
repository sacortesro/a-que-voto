package controllers

import (
	"net/http"
	"strconv"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
)

func (apiCfg *APIConfig) HandleGetRawDocuments(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("source_id"), 10, 16)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid source")
		return
	}

	documentList, err := apiCfg.Store.Queries().ListRawDocumentsBySource(r.Context(), int16(id))
	if err != nil {
		utils.RespondWithInternalError(w, "get raw document", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.DatabaseRawDocumentToRawDocument(documentList))
}
