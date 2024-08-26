package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"log"
	"main/modules/userFiles"
	"net/http"
	"os"
)

func setupRoutes(mainRouter *chi.Mux) {
	userFilesRouter := chi.NewRouter()
	userFilesRouter.Post("/upload", userFiles.CreateFile)
	userFilesRouter.Get("/download", userFiles.SelectFile)
	mainRouter.Route("/", func(r chi.Router) {
		mainRouter.Mount("/user_files", userFilesRouter)
	})
}

func main() {
	godotenv.Load(".env." + os.Getenv("APP_ENV"))
	r := chi.NewRouter()
	r.Use(middleware.AllowContentType("application/json", "multipart/form-data"))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	setupRoutes(r)

	httpServer := &http.Server{Addr: "localhost:8000", Handler: r}
	err := httpServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf(err.Error())
	}
}
