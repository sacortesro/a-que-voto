package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/database"
)

var emptySocialMedia = json.RawMessage(`[]`)

func (apiCfg *APIConfig) HandleCreatePerson(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		FullName    string          `json:"full_name"`
		BirthDate   *string         `json:"birth_date"`
		PhotoURL    *string         `json:"photo_url"`
		SocialMedia json.RawMessage `json:"social_media"`
	}

	decoder := json.NewDecoder(r.Body)

	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	birthDate, err := utils.ParseOptionalDate(params.BirthDate)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "birth_date must be formatted as YYYY-MM-DD")
		return
	}

	socialMedia := params.SocialMedia
	if len(socialMedia) == 0 {
		socialMedia = emptySocialMedia
	}

	ctx := r.Context()

	var person database.Person
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		person, err = q.CreatePerson(ctx, database.CreatePersonParams{
			FullName:    params.FullName,
			BirthDate:   utils.NullTime(birthDate),
			PhotoUrl:    utils.NullString(params.PhotoURL),
			SocialMedia: socialMedia,
		})
		return err
	})
	if utils.IsCheckViolation(err) {
		utils.RespondWithError(w, http.StatusBadRequest, "social_media is invalid")
		return
	}
	if err != nil {
		utils.RespondWithInternalError(w, "create person", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, models.DatabasePersonToPerson(person))
}
