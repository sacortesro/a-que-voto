package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

func (apiCfg *APIConfig) HandleGetPerson(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid person id")
		return
	}

	person, err := apiCfg.Store.Queries().GetPersonByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.RespondWithError(w, http.StatusNotFound, "person not found")
			return
		}
		utils.RespondWithInternalError(w, "get person", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.DatabasePersonToPerson(person))
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

func (apiCfg *APIConfig) HandleUpdatePerson(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid person id")
		return
	}

	type parameters struct {
		FullName          string          `json:"full_name"`
		BirthDate         *string         `json:"birth_date"`
		PhotoURL          *string         `json:"photo_url"`
		SocialMedia       json.RawMessage `json:"social_media"`
		IDCongresoVisible *int32          `json:"id_congreso_visible"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	fullName := strings.TrimSpace(params.FullName)
	if fullName == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "mandatory full_name")
		return
	}

	birthDate, err := utils.ParseOptionalDate(params.BirthDate)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "birth_date must be formatted as YYYY-MM-DD")
		return
	}

	socialMedia := params.SocialMedia
	if len(socialMedia) == 0 || string(socialMedia) == "null" {
		socialMedia = emptySocialMedia
	}

	if params.IDCongresoVisible != nil && *params.IDCongresoVisible <= 0 {
		utils.RespondWithError(w, http.StatusBadRequest, "id_congreso_visible must be greater than zero")
		return
	}

	var person database.Person
	err = apiCfg.Store.Write(r.Context(), func(q *database.Queries) error {
		var err error
		person, err = q.UpdatePerson(r.Context(), database.UpdatePersonParams{
			ID:                id,
			FullName:          fullName,
			BirthDate:         utils.NullTime(birthDate),
			PhotoUrl:          utils.NullString(params.PhotoURL),
			SocialMedia:       socialMedia,
			IDCongresoVisible: utils.NullInt32(params.IDCongresoVisible),
		})
		return err
	})
	if utils.IsCheckViolation(err) {
		utils.RespondWithError(w, http.StatusBadRequest, "social_media is invalid")
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		utils.RespondWithError(w, http.StatusNotFound, "person not found")
		return
	}
	if err != nil {
		utils.RespondWithInternalError(w, "update person", err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, models.DatabasePersonToPerson(person))
}
