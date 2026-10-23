package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/sacortesro/a-que-voto/cmd/app/controllers"
	"github.com/sacortesro/a-que-voto/internal/auth"
)

func GetRouter(cfg *controllers.APIConfig, resolve auth.Resolver) *chi.Mux {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.Use(auth.Middleware(resolve))

	router.Route("/v1", func(r chi.Router) {
		r.Get("/healthz", controllers.HandleKeepAlive)
		r.Post("/source", cfg.HandleCreateSource)
		r.Post("/term", cfg.HandleCreatePeriod)
		r.Post("/party", cfg.HandleCreateParty)
		r.Post("/district", cfg.HandleCreateDistrict)
		r.Post("/people", cfg.HandleCreatePerson)
	})

	return router
}
