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

func DatabaseSourcesToSources(dbSourceList []database.Source) []Source {
	sourceList := make([]Source, 0, len(dbSourceList))
	for _, source := range dbSourceList {
		sourceList = append(sourceList,
			Source{
				ID:        source.ID,
				Code:      source.Code,
				BaseURL:   source.BaseUrl,
				Priority:  source.Priority,
				CreatedAt: utils.FormatTimestamp(source.CreatedAt),
				UpdatedAt: utils.FormatTimestamp(source.UpdatedAt),
			})
	}
	return sourceList
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

func DatabaseTermsToTerms(dbTerms []database.Term) []Term {
	termList := make([]Term, 0, len(dbTerms))
	for _, term := range dbTerms {
		termList = append(termList, Term{
			ID:        term.ID,
			Label:     term.Label,
			StartDate: utils.FormatDate(term.StartDate),
			EndDate:   utils.FormatDate(term.EndDate),
			CreatedAt: utils.FormatTimestamp(term.CreatedAt),
			UpdatedAt: utils.FormatTimestamp(term.UpdatedAt),
		})
	}
	return termList
}

type Party struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Acronym   string    `json:"acronym"`
	IsActive  bool      `json:"is_active"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

func DatabasePartyToParty(dbParty database.Party) Party {
	return Party{
		ID:        dbParty.ID,
		Name:      dbParty.Name,
		Acronym:   dbParty.Acronym.String,
		IsActive:  dbParty.IsActive,
		CreatedAt: utils.FormatTimestamp(dbParty.CreatedAt),
		UpdatedAt: utils.FormatTimestamp(dbParty.UpdatedAt),
	}
}

func DatabasePartiesToParties(dbParties []database.Party) []Party {
	partyList := make([]Party, 0, len(dbParties))
	for _, party := range dbParties {
		partyList = append(partyList, DatabasePartyToParty(party))
	}
	return partyList
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

func DatabaseDistrictsToDistricts(dbDistricts []database.District) []District {
	districtList := make([]District, 0, len(dbDistricts))
	for _, district := range dbDistricts {
		districtList = append(districtList, DatabaseDistrictToDistrict(district))
	}
	return districtList
}

type Person struct {
	ID                uuid.UUID       `json:"id"`
	FullName          string          `json:"full_name"`
	NameNorm          string          `json:"name_norm"`
	BirthDate         string          `json:"birth_date"`
	PhotoURL          string          `json:"photo_url"`
	SocialMedia       json.RawMessage `json:"social_media"`
	IDCongresoVisible *int32          `json:"id_congreso_visible"`
	CreatedAt         string          `json:"created_at"`
	UpdatedAt         string          `json:"updated_at"`
}

func DatabasePersonToPerson(person database.Person) Person {
	var idCongresoVisible *int32
	if person.IDCongresoVisible.Valid {
		v := person.IDCongresoVisible.Int32
		idCongresoVisible = &v
	}

	return Person{
		ID:                person.ID,
		FullName:          person.FullName,
		NameNorm:          person.NameNorm.String,
		BirthDate:         utils.FormatNullDate(person.BirthDate),
		PhotoURL:          person.PhotoUrl.String,
		SocialMedia:       person.SocialMedia,
		IDCongresoVisible: idCongresoVisible,
		CreatedAt:         utils.FormatTimestamp(person.CreatedAt),
		UpdatedAt:         utils.FormatTimestamp(person.UpdatedAt),
	}
}

func DatabasePersonsToPersons(dbPersons []database.Person) []Person {
	personList := make([]Person, 0, len(dbPersons))
	for _, person := range dbPersons {
		personList = append(personList, DatabasePersonToPerson(person))
	}
	return personList
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

func DatabaseMandatesToMandates(dbMandates []database.Mandate) []Mandate {
	mandateList := make([]Mandate, 0, len(dbMandates))
	for _, mandate := range dbMandates {
		mandateList = append(mandateList, DatabaseMandateToMandate(mandate))
	}
	return mandateList
}

type Committee struct {
	ID        uuid.UUID `json:"id"`
	Chamber   string    `json:"chamber"`
	Kind      string    `json:"kind"`
	Name      string    `json:"name"`
	Number    int16     `json:"number"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

func DatabaseCommitteeToCommittee(dbCommittee database.Committee) Committee {
	return Committee{
		ID:        dbCommittee.ID,
		Chamber:   dbCommittee.Chamber,
		Kind:      dbCommittee.Kind,
		Name:      dbCommittee.Name,
		Number:    dbCommittee.Number.Int16,
		CreatedAt: utils.FormatTimestamp(dbCommittee.CreatedAt),
		UpdatedAt: utils.FormatTimestamp(dbCommittee.UpdatedAt),
	}
}

type CommitteeMembership struct {
	MandateID   uuid.UUID `json:"mandate_id"`
	CommitteeID uuid.UUID `json:"committee_id"`
	Role        string    `json:"role"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

func DatabaseCommitteeMembershipToCommitteeMembership(dbMembership database.CommitteeMembership) CommitteeMembership {
	return CommitteeMembership{
		MandateID:   dbMembership.MandateID,
		CommitteeID: dbMembership.CommitteeID,
		Role:        dbMembership.Role,
		StartDate:   utils.FormatDate(dbMembership.StartDate),
		EndDate:     utils.FormatNullDate(dbMembership.EndDate),
		CreatedAt:   utils.FormatTimestamp(dbMembership.CreatedAt),
		UpdatedAt:   utils.FormatTimestamp(dbMembership.UpdatedAt),
	}
}

type Stage struct {
	Slug      string `json:"slug"`
	Label     string `json:"label"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func DatabaseStagesToStages(dbStageList []database.Stage) []Stage {
	stageList := make([]Stage, 0, len(dbStageList))
	for _, stage := range dbStageList {
		stageList = append(stageList,
			Stage{
				Slug:      stage.Slug,
				Label:     stage.Label,
				CreatedAt: utils.FormatTimestamp(stage.CreatedAt),
				UpdatedAt: utils.FormatTimestamp(stage.UpdatedAt),
			})
	}
	return stageList
}

type DocumentKind struct {
	Slug      string `json:"slug"`
	Label     string `json:"label"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func DatabaseDocumentKindsToDocumentKinds(dbDocumentKinds []database.DocumentKind) []DocumentKind {
	kindList := make([]DocumentKind, 0, len(dbDocumentKinds))
	for _, kind := range dbDocumentKinds {
		kindList = append(kindList,
			DocumentKind{
				Slug:      kind.Slug,
				Label:     kind.Label,
				CreatedAt: utils.FormatTimestamp(kind.CreatedAt),
				UpdatedAt: utils.FormatTimestamp(kind.UpdatedAt),
			})
	}
	return kindList
}

type RawDocument struct {
	ID          uuid.UUID `json:"id"`
	SourceID    int16     `json:"source_id"`
	URL         string    `json:"url"`
	ContentType string    `json:"content_type"`
	FetchedAt   string    `json:"fetched_at"`
	FetchRunID  uuid.UUID `json:"fetch_run_id"`
	CreatedAt   string    `json:"created_at"`
}

func DatabaseRawDocumentToRawDocument(dbDocuments []database.RawDocument) []RawDocument {
	documentList := make([]RawDocument, 0, len(dbDocuments))
	for _, doc := range dbDocuments {
		documentList = append(documentList, RawDocument{
			ID:          doc.ID,
			SourceID:    doc.SourceID,
			URL:         doc.Url,
			ContentType: doc.ContentType.String,
			FetchRunID:  doc.FetchRunID.UUID,
			FetchedAt:   utils.FormatTimestamp(doc.FetchedAt),
			CreatedAt:   utils.FormatTimestamp(doc.CreatedAt),
		})
	}
	return documentList
}
