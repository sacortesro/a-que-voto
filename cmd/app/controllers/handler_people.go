package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sacortesro/a-que-voto/cmd/app/models"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/database"
)

const (
	defaultRows = 50
	maxRows     = 1000
	defaultPage = 1
	maxPage     = 100
)

func (apiCfg *APIConfig) HandleListPersons(w http.ResponseWriter, r *http.Request) {

	rows, err := parsePositiveInt(r, "rows", defaultRows, maxRows)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	page, err := parsePositiveInt(r, "page", defaultPage, maxPage)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	persons, err := apiCfg.Store.Queries().ListPersons(r.Context(), database.ListPersonsParams{
		Limit:  int32(rows),
		Offset: int32(rows * (page - 1)),
	})

	if err != nil {
		utils.RespondWithInternalError(w, "list persons", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.DatabasePersonsToPersons(persons))
}

func parsePositiveInt(r *http.Request, key string, def int, max int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > max {
		return 0, fmt.Errorf("%s must be an integer between 1 and %d", key, max)
	}
	return n, nil
}

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
