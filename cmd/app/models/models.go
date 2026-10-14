package models

import "github.com/sacortesro/a-que-voto/internal/database"

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
