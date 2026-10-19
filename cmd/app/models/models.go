package models

import (
	"time"

	"github.com/sacortesro/a-que-voto/internal/database"
)

type Source struct {
	ID        int16  `json:"id"`
	Code      string `json:"code"`
	BaseUrl   string `json:"base_url"`
	Priority  int16  `json:"priority"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func DatabaseSourceToSource(dbSource database.Source) Source {
	return Source{
		ID:        dbSource.ID,
		Code:      dbSource.Code,
		BaseUrl:   dbSource.BaseUrl,
		Priority:  dbSource.Priority,
		CreatedAt: dbSource.CreatedAt.String(),
		UpdatedAt: dbSource.UpdatedAt.String(),
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
