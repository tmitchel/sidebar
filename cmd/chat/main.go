package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/server"
	"github.com/tmitchel/sidebar/services"
	"github.com/tmitchel/sidebar/store"
)

func main() {
	// load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		logrus.Error("Error loading .env file. If not deploying, consider checking.")
	}

	// read database connection info from env variables
	var dbConn string
	if os.Getenv("PRODDY") == "true" {
		dbConn = os.Getenv("DATABASE_URL")
	} else {
		dbConn = fmt.Sprintf("host=db port=5432 user=%s "+
			"password=%s dbname=%s sslmode=disable",
			os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))
	}

	// open the database connection and defer closing it
	db, err := store.New(dbConn)
	if err != nil {
		logrus.Fatal(err)
	}
	defer db.Close()

	// setup all services
	auth, err := services.NewAuthenticater(db)
	if err != nil {
		logrus.Fatal(err)
	}

	create, err := services.NewCreater(db)
	if err != nil {
		logrus.Fatal(err)
	}

	delete, err := services.NewDeleter(db)
	if err != nil {
		logrus.Fatal(err)
	}

	add, err := services.NewAdder(db)
	if err != nil {
		logrus.Fatal(err)
	}

	get, err := services.NewGetter(db)
	if err != nil {
		logrus.Fatal(err)
	}

	up, err := services.NewUpdater(db)
	if err != nil {
		logrus.Fatal(err)
	}

	// if the database is empty, create a default workspace
	if err := db.Empty(); err != nil {
		db.CreateWorkspace(&sidebar.Workspace{
			ID:          uuid.New().String(),
			DisplayName: os.Getenv("DEFAULT_DISPLAYNAME"),
			Token:       os.Getenv("DEFAULT_TOKEN"),
		})
	}

	// build the server and inject dependencies
	server := server.NewServer(auth, create, delete, add, get, up)

	// serve
	if os.Getenv("PORT") != "" {
		http.ListenAndServe(":"+os.Getenv("PORT"), server.Serve())
	} else {
		http.ListenAndServeTLS(":8080", "localhost.pem", "localhost-key.pem", server.Serve())
	}
}
