package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
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

	rows, err := utils.ParseQueryPositiveInt(r, "rows", defaultRows, maxRows)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	page, err := utils.ParseQueryPositiveInt(r, "page", defaultPage, maxPage)
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

type personInput struct {
	FullName          string
	BirthDate         sql.NullTime
	PhotoURL          sql.NullString
	SocialMedia       json.RawMessage
	IDCongresoVisible sql.NullInt32
}

func decodePersonInput(r *http.Request) (personInput, error) {
	var params struct {
		FullName          string          `json:"full_name"`
		BirthDate         *string         `json:"birth_date"`
		PhotoURL          *string         `json:"photo_url"`
		SocialMedia       json.RawMessage `json:"social_media"`
		IDCongresoVisible *int32          `json:"id_congreso_visible"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&params); err != nil {
		return personInput{}, errors.New("invalid JSON body")
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return personInput{}, errors.New("invalid JSON body")
	}

	fullName := strings.TrimSpace(params.FullName)
	if fullName == "" {
		return personInput{}, errors.New("mandatory full name")
	}

	birthDate, err := utils.ParseOptionalDate(params.BirthDate)
	if err != nil {
		return personInput{}, errors.New("birth_date must be formatted as YYYY-MM-DD")
	}

	var photoURL *string
	if params.PhotoURL != nil {
		trimmed := strings.TrimSpace(*params.PhotoURL)
		photoURL = &trimmed
	}

	socialMedia := params.SocialMedia
	if len(socialMedia) == 0 || string(socialMedia) == "null" {
		socialMedia = emptySocialMedia
	}

	if params.IDCongresoVisible != nil && *params.IDCongresoVisible <= 0 {
		return personInput{}, errors.New("id_congreso_visible must be greater than zero")
	}

	return personInput{
		FullName:          fullName,
		BirthDate:         utils.NullTime(birthDate),
		PhotoURL:          utils.NullString(photoURL),
		SocialMedia:       socialMedia,
		IDCongresoVisible: utils.NullInt32(params.IDCongresoVisible),
	}, nil
}

func (apiCfg *APIConfig) HandleCreatePerson(w http.ResponseWriter, r *http.Request) {
	input, err := decodePersonInput(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	var person database.Person
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		person, err = q.CreatePerson(ctx, database.CreatePersonParams{
			FullName:          input.FullName,
			BirthDate:         input.BirthDate,
			PhotoUrl:          input.PhotoURL,
			SocialMedia:       input.SocialMedia,
			IDCongresoVisible: input.IDCongresoVisible,
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

	input, err := decodePersonInput(r)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	ctx := r.Context()

	var person database.Person
	err = apiCfg.Store.Write(ctx, func(q *database.Queries) error {
		var err error
		person, err = q.UpdatePerson(ctx, database.UpdatePersonParams{
			ID:                id,
			FullName:          input.FullName,
			BirthDate:         input.BirthDate,
			PhotoUrl:          input.PhotoURL,
			SocialMedia:       input.SocialMedia,
			IDCongresoVisible: input.IDCongresoVisible,
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
