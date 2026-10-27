package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

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

	q := apiCfg.Store.Queries()

	_, err = q.GetSourceByCode(ctx, params.Code)
	if err == nil {
		utils.RespondWithError(w, http.StatusConflict, "code already exists")
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		utils.RespondWithInternalError(w, "get source", err)
		return
	}

	validURL, ok := validBaseURL(params.BaseURL)
	if !ok {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid URL. base_url must be https URL")
		return
	}

	var source database.Source
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		source, err = q.CreateSource(ctx, database.CreateSourceParams{
			Code:     params.Code,
			BaseUrl:  validURL,
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

func validBaseURL(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", false
	}
	return strings.TrimSuffix(u.String(), "/"), true
}
