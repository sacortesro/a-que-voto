package controllers

import (
	"github.com/sacortesro/a-que-voto/internal/store"
)

type APIConfig struct {
	Store *store.Store
}

func NewAPIConfig(s *store.Store) *APIConfig {
	return &APIConfig{
		Store: s,
	}
}
