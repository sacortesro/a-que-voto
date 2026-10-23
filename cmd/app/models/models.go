package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/database"
)

type Source struct {
	ID        int16     `json:"id"`
	Code      string    `json:"code"`
	BaseURL   string    `json:"base_url"`
	Priority  int16     `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func DatabaseSourceToSource(dbSource database.Source) Source {
	return Source{
		ID:        dbSource.ID,
		Code:      dbSource.Code,
		BaseURL:   dbSource.BaseUrl,
		Priority:  dbSource.Priority,
		CreatedAt: dbSource.CreatedAt,
		UpdatedAt: dbSource.UpdatedAt,
	}
}

type Term struct {
	ID        int16     `json:"id"`
	Label     string    `json:"label"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func DatabaseTermToTerm(dbTerm database.Term) Term {
	return Term{
		ID:        dbTerm.ID,
		Label:     dbTerm.Label,
		StartDate: dbTerm.StartDate,
		EndDate:   dbTerm.EndDate,
		CreatedAt: dbTerm.CreatedAt,
		UpdatedAt: dbTerm.UpdatedAt,
	}
}

type Party struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Acronym   string    `json:"acronym"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func DatabasePartyToParty(dbParty database.Party) Party {
	return Party{
		ID:        dbParty.ID,
		Name:      dbParty.Name,
		Acronym:   dbParty.Acronym.String,
		CreatedAt: dbParty.CreatedAt,
		UpdatedAt: dbParty.UpdatedAt,
	}
}

type District struct {
	ID         int16     `json:"id"`
	Chamber    string    `json:"chamber"`
	Kind       string    `json:"kind"`
	Department string    `json:"department"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func DatabaseDistrictToDistrict(dbDistrict database.District) District {
	return District{
		ID:         dbDistrict.ID,
		Chamber:    dbDistrict.Chamber,
		Kind:       dbDistrict.Kind,
		Department: dbDistrict.Department.String,
		Name:       dbDistrict.Name,
		CreatedAt:  dbDistrict.CreatedAt,
		UpdatedAt:  dbDistrict.UpdatedAt,
	}
}

type Person struct {
	ID          uuid.UUID       `json:"id"`
	FullName    string          `json:"full_name"`
	NameNorm    string          `json:"name_norm"`
	BirthDate   string          `json:"birth_date"`
	PhotoURL    string          `json:"photo_url"`
	SocialMedia json.RawMessage `json:"social_media"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func DatabasePersonToPerson(person database.Person) Person {
	return Person{
		ID:          person.ID,
		FullName:    person.FullName,
		NameNorm:    person.NameNorm.String,
		BirthDate:   utils.FormatNullDate(person.BirthDate),
		PhotoURL:    person.PhotoUrl.String,
		SocialMedia: person.SocialMedia,
		CreatedAt:   person.CreatedAt,
		UpdatedAt:   person.UpdatedAt,
	}
}
