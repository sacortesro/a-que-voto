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

	router.Route("/v1", func(r chi.Router) {
		r.Get("/healthz", controllers.HandleKeepAlive)

		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(resolve))

			r.Get("/stage", cfg.HandleGetAllStages)
			r.Get("/document_kind", cfg.HandleGetAllDocumentKinds)
			r.Get("/source", cfg.HandleGetAllSources)

			r.Post("/term", cfg.HandleCreatePeriod)
			r.Post("/party", cfg.HandleCreateParty)
			r.Post("/district", cfg.HandleCreateDistrict)
			r.Post("/people", cfg.HandleCreatePerson)
			r.Post("/mandate", cfg.HandlerCreateMandate)
			r.Post("/committee", cfg.HandleCreateCommittee)
			r.Post("/committee_membership", cfg.HandleAddPersonToCommittee)

			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(auth.RoleAdmin))

				r.Post("/source", cfg.HandleCreateSource)
			})
		})
	})

	return router
}
