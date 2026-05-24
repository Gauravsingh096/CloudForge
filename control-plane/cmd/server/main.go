package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/Gauravsingh096/cloudforge/control-plane/internal/api"
	"github.com/Gauravsingh096/cloudforge/control-plane/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()

	database, err := db.Connect(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer database.Close()

	r := router(database)

	addr := ":" + port()
	log.Printf("cloudforge control-plane listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func router(database *sql.DB) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	h := api.NewHandler(database)

	r.Get("/healthz", h.Healthz)

	r.Route("/api", func(r chi.Router) {
		r.Get("/projects", h.ListProjects)
		r.Post("/projects", h.CreateProject)
		r.Get("/projects/{id}", h.GetProject)
		r.Get("/projects/{id}/deploys", h.ListDeploys)

		r.Post("/deploys", h.CreateDeploy)
		r.Get("/deploys/{id}", h.GetDeploy)

		r.Post("/webhook/github", h.GithubWebhook)
	})

	return r
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
