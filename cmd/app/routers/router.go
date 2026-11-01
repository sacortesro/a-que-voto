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
			// vocabularies
			r.Get("/stage", cfg.HandleGetAllStages)
			r.Get("/document_kind", cfg.HandleGetAllDocumentKinds)
			// sources
			r.Get("/source", cfg.HandleGetAllSources)
			r.Get("/raw_document", cfg.HandleGetRawDocuments)
			// term
			r.Get("/term", cfg.HandleGetTerms)
			// party
			r.Get("/party", cfg.HandleGetParties)
			// district
			r.Get("/district", cfg.HandleListDistricts)
			// person
			r.Get("/people", cfg.HandleListPersons)
			r.Get("/people/{id}", cfg.HandleGetPerson)
			// mandate
			r.Get("/people/{id}/mandates", cfg.HandlerGetMandatesByPerson)
			// committee
			r.Get("/committee", cfg.HandleListCommittees)
			r.Post("/committee_membership", cfg.HandleAddPersonToCommittee)

			r.Group(func(r chi.Router) {
				r.Use(auth.RequireRole(auth.RoleAdmin))

				r.Post("/source", cfg.HandleCreateSource)
				r.Put("/source/{id}", cfg.HandleUpdateSource)
				// periods
				r.Post("/term", cfg.HandleCreatePeriod)
				// party
				r.Post("/party", cfg.HandleCreateParty)
				r.Put("/party/{id}", cfg.HandleUpdateParty)
				// district
				r.Post("/district", cfg.HandleCreateDistrict)
				// person
				r.Post("/people", cfg.HandleCreatePerson)
				r.Put("/people/{id}", cfg.HandleUpdatePerson)
				// mandate
				r.Post("/mandate", cfg.HandlerCreateMandate)
				// committee
				r.Post("/committee", cfg.HandleCreateCommittee)
				r.Put("/committee/{id}", cfg.HandleUpdateCommittee)
			})
		})
	})

	return router
}
