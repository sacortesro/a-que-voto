package models

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sacortesro/a-que-voto/cmd/app/utils"
	"github.com/sacortesro/a-que-voto/internal/database"
)

type Source struct {
	ID        int16  `json:"id"`
	Code      string `json:"code"`
	BaseURL   string `json:"base_url"`
	Priority  int16  `json:"priority"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func DatabaseSourceToSource(dbSource database.Source) Source {
	return Source{
		ID:        dbSource.ID,
		Code:      dbSource.Code,
		BaseURL:   dbSource.BaseUrl,
		Priority:  dbSource.Priority,
		CreatedAt: utils.FormatTimestamp(dbSource.CreatedAt),
		UpdatedAt: utils.FormatTimestamp(dbSource.UpdatedAt),
	}
}

type Term struct {
	ID        int16  `json:"id"`
	Label     string `json:"label"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func DatabaseTermToTerm(dbTerm database.Term) Term {
	return Term{
		ID:        dbTerm.ID,
		Label:     dbTerm.Label,
		StartDate: utils.FormatDate(dbTerm.StartDate),
		EndDate:   utils.FormatDate(dbTerm.EndDate),
		CreatedAt: utils.FormatTimestamp(dbTerm.CreatedAt),
		UpdatedAt: utils.FormatTimestamp(dbTerm.UpdatedAt),
	}
}

type Party struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Acronym   string    `json:"acronym"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

func DatabasePartyToParty(dbParty database.Party) Party {
	return Party{
		ID:        dbParty.ID,
		Name:      dbParty.Name,
		Acronym:   dbParty.Acronym.String,
		CreatedAt: utils.FormatTimestamp(dbParty.CreatedAt),
		UpdatedAt: utils.FormatTimestamp(dbParty.UpdatedAt),
	}
}

type District struct {
	ID         int16  `json:"id"`
	Chamber    string `json:"chamber"`
	Kind       string `json:"kind"`
	Department string `json:"department"`
	Name       string `json:"name"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func DatabaseDistrictToDistrict(dbDistrict database.District) District {
	return District{
		ID:         dbDistrict.ID,
		Chamber:    dbDistrict.Chamber,
		Kind:       dbDistrict.Kind,
		Department: dbDistrict.Department.String,
		Name:       dbDistrict.Name,
		CreatedAt:  utils.FormatTimestamp(dbDistrict.CreatedAt),
		UpdatedAt:  utils.FormatTimestamp(dbDistrict.UpdatedAt),
	}
}

type Person struct {
	ID          uuid.UUID       `json:"id"`
	FullName    string          `json:"full_name"`
	NameNorm    string          `json:"name_norm"`
	BirthDate   string          `json:"birth_date"`
	PhotoURL    string          `json:"photo_url"`
	SocialMedia json.RawMessage `json:"social_media"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

func DatabasePersonToPerson(person database.Person) Person {
	return Person{
		ID:          person.ID,
		FullName:    person.FullName,
		NameNorm:    person.NameNorm.String,
		BirthDate:   utils.FormatNullDate(person.BirthDate),
		PhotoURL:    person.PhotoUrl.String,
		SocialMedia: person.SocialMedia,
		CreatedAt:   utils.FormatTimestamp(person.CreatedAt),
		UpdatedAt:   utils.FormatTimestamp(person.UpdatedAt),
	}
}

type Mandate struct {
	ID         uuid.UUID `json:"id"`
	PersonID   uuid.UUID `json:"person_id"`
	TermID     int16     `json:"term_id"`
	Chamber    string    `json:"chamber"`
	DistrictID int16     `json:"district_id"`
	PartyID    uuid.UUID `json:"party_id"`
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
	Status     string    `json:"status"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
}

func DatabaseMandateToMandate(dbMandate database.Mandate) Mandate {
	return Mandate{
		ID:         dbMandate.ID,
		PersonID:   dbMandate.PersonID,
		TermID:     dbMandate.TermID,
		Chamber:    dbMandate.Chamber,
		DistrictID: dbMandate.DistrictID.Int16,
		PartyID:    dbMandate.PartyID.UUID,
		StartDate:  utils.FormatDate(dbMandate.StartDate),
		EndDate:    utils.FormatNullDate(dbMandate.EndDate),
		Status:     dbMandate.Status,
		CreatedAt:  utils.FormatTimestamp(dbMandate.CreatedAt),
		UpdatedAt:  utils.FormatTimestamp(dbMandate.UpdatedAt),
	}
}
