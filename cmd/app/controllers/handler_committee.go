package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/database"
)

const (
	committeeConstitutional = "constitucional"
	committeeLegal          = "legal"
	committeeAccidental     = "accidental"
)

func (apiCfg *APIConfig) HandleCreateCommittee(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Chamber string `json:"chamber"`
		Kind    string `json:"kind"`
		Name    string `json:"name"`
		Number  *int16 `json:"number"`
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

	if !isValidCommittee(params.Kind) {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid committee kind")
		return
	}

	params.Name = strings.TrimSpace(params.Name)
	if params.Name == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid committee name")
		return
	}

	ctx := r.Context()
	var committee database.Committee
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		committee, err = q.CreateCommittee(ctx, database.CreateCommitteeParams{
			Chamber: params.Chamber,
			Kind:    params.Kind,
			Name:    params.Name,
			Number:  utils.NullInt16(params.Number),
		})
		return err
	})
	if err != nil {
		utils.RespondWithInternalError(w, "create committee", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, models.DatabaseCommitteeToCommittee(committee))
}

func isValidCommittee(committee string) bool {
	return committee == committeeAccidental ||
		committee == committeeConstitutional ||
		committee == committeeLegal
}
