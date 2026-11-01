package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/database"
)

const (
	committeeConstitutional    = "constitucional"
	committeeLegal             = "legal"
	committeeAccidental        = "accidental"
	committeeRoleMember        = "miembro"
	committeeRolePresident     = "presidente"
	committeeRoleVicepresident = "vicepresidente"
)

func (apiCfg *APIConfig) HandleListCommittees(w http.ResponseWriter, r *http.Request) {
	committees, err := apiCfg.Store.Queries().ListCommittees(r.Context())
	if err != nil {
		utils.RespondWithInternalError(w, "list committees", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.DatabaseCommitteesToCommittees(committees))
}

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

func (apiCfg *APIConfig) HandleAddPersonToCommittee(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		MandateID   uuid.UUID `json:"mandate_id"`
		CommitteeID uuid.UUID `json:"committee_id"`
		Role        string    `json:"role"`
		StartDate   string    `json:"start_date"`
		EndDate     *string   `json:"end_date"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if !isValidCommitteeRole(params.Role) {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid committee role")
		return
	}

	startDate, err := utils.ParseDate(params.StartDate)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "start_date must be formatted as YYYY-MM-DD")
		return
	}

	endDate, err := utils.ParseOptionalDate(params.EndDate)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "end_date must be formatted as YYYY-MM-DD")
		return
	}
	if endDate != nil && endDate.Before(startDate) {
		utils.RespondWithError(w, http.StatusBadRequest, "end_date must not be before start_date")
		return
	}

	ctx := r.Context()
	q := apiCfg.Store.Queries()

	_, err = q.GetMandateByID(ctx, params.MandateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, http.StatusBadRequest, "mandate not found")
			return
		}
		utils.RespondWithInternalError(w, "get mandate", err)
		return
	}

	_, err = q.GetCommitteeByID(ctx, params.CommitteeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, http.StatusBadRequest, "committee not found")
			return
		}
		utils.RespondWithInternalError(w, "get committee", err)
		return
	}

	_, err = q.GetCommitteeMembership(ctx, database.GetCommitteeMembershipParams{
		MandateID:   params.MandateID,
		CommitteeID: params.CommitteeID,
		StartDate:   startDate,
	})
	if err == nil {
		utils.RespondWithError(w, http.StatusConflict, "committee membership already exists")
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		utils.RespondWithInternalError(w, "get committee membership", err)
		return
	}

	var committeeMembership database.CommitteeMembership
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		committeeMembership, err = q.InsertNewCommitteeMember(ctx, database.InsertNewCommitteeMemberParams{
			MandateID:   params.MandateID,
			CommitteeID: params.CommitteeID,
			Role:        params.Role,
			StartDate:   startDate,
			EndDate:     utils.NullTime(endDate),
		})
		return err
	})
	if err != nil {
		utils.RespondWithInternalError(w, "create committee membership", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, models.DatabaseCommitteeMembershipToCommitteeMembership(committeeMembership))
}

func isValidCommitteeRole(role string) bool {
	return role == committeeRoleMember ||
		role == committeeRolePresident ||
		role == committeeRoleVicepresident
}

func isValidCommittee(committee string) bool {
	return committee == committeeAccidental ||
		committee == committeeConstitutional ||
		committee == committeeLegal
}
