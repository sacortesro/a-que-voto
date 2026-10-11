package controllers

import "github.com/sacortesro/a-que-voto/internal/database"

type apiConfig struct {
	DB *database.Queries
}

func NewAPIConfig(db *database.Queries) *apiConfig {
	return &apiConfig{DB: db}
}
