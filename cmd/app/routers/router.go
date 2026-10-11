package routers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/sacortesro/a-que-voto/cmd/app/controllers"
)

func GetRouter() *chi.Mux {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router1 := chi.NewRouter()
	router1.Get("/healthz", controllers.HandleKeepAlive)

	router.Mount("/v1", router1)

	return router
}
