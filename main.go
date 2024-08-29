package main

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"log"
	"main/modules/common"
	"main/modules/database"
	"main/modules/userFiles"
	"net/http"
	"os"
	"time"
)

func initConfig(c chan<- string) {
	godotenv.Load(".env." + os.Getenv("APP_ENV"))
	common.Config = common.NewConfig()
	c <- "config"
	fmt.Println("Config initialized")
}

func initAfter(c <-chan string) {
	for {
		msg := <-c
		if msg == "config" {
			database.DbConfig = database.NewDatabase(
				common.Config.Get("DB_HOST"),
				common.Config.Get("DB_PORT"),
				common.Config.Get("DB_USER"),
				common.Config.Get("DB_PASSWORD"),
				common.Config.Get("DB_DATABASE"),
			)
			database.Redis = database.NewRedisService()
			userFiles.Service = &userFiles.UserFilesService{}
			break
		}
		time.Sleep(time.Second * 1)
	}
	fmt.Println("Init completed")
}

func initDependencies() {
	c := make(chan string)
	go initConfig(c)
	go initAfter(c)
}

func setupRoutes(mainRouter *chi.Mux) {
	userFilesRouter := chi.NewRouter()
	userFilesRouter.Post("/upload", userFiles.CreateFile)
	userFilesRouter.Get("/download", userFiles.SelectFile)
	userFilesRouter.Get("/", userFiles.GetUserFilesList)
	userFilesRouter.Patch("/", userFiles.PatchUserFile)
	userFilesRouter.Delete("/", userFiles.DeleteUserFile)
	mainRouter.Route("/", func(r chi.Router) {
		mainRouter.Mount("/user_files", userFilesRouter)
	})
}

func main() {
	initDependencies()
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
