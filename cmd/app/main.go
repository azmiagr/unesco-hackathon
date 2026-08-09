package main

import (
	"log"

	"github.com/azmiagr/unesco-hackathon/internal/handler/rest"
	"github.com/azmiagr/unesco-hackathon/internal/repository"
	"github.com/azmiagr/unesco-hackathon/internal/service"
	"github.com/azmiagr/unesco-hackathon/pkg/bcrypt"
	"github.com/azmiagr/unesco-hackathon/pkg/config"
	"github.com/azmiagr/unesco-hackathon/pkg/database/mariadb"
	"github.com/azmiagr/unesco-hackathon/pkg/jwt"
	"github.com/azmiagr/unesco-hackathon/pkg/middleware"
	"github.com/azmiagr/unesco-hackathon/pkg/supabase"
)

func main() {
	config.LoadEnvironment()

	db, err := mariadb.ConnectDatabase()
	if err != nil {
		log.Fatal(err)
	}

	err = mariadb.Migrate(db)
	if err != nil {
		log.Fatal(err)
	}

	err = mariadb.Seed(db)
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewRepository(db)
	bcrypt := bcrypt.Init()
	jwt := jwt.Init()
	supabase := supabase.Init()
	svc := service.NewService(repo, bcrypt, jwt, supabase)

	middleware := middleware.Init(svc, jwt)
	r := rest.NewRest(svc, middleware)
	r.MountEndpoint()

	r.Run()
}
