package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/services"
	"github.com/tmitchel/sidebar/store"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Error("Error loading .env file. If not deploying, consider checking.")
	}

	var dbConn string
	if os.Getenv("PRODDY") == "true" {
		dbConn = os.Getenv("DATABASE_URL")
	} else {
		dbConn = fmt.Sprintf("host=localhost port=5432 user=%s "+
			"password=%s dbname=%s sslmode=disable",
			os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))
	}
	db, err := store.New(dbConn)
	if err != nil {
		logrus.Fatal(err)
	}
	defer db.Close()

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

	if u, err := get.GetUsers(); err != nil {
		logrus.Fatal("Can't query for users on start")
	} else if len(u) == 0 {
		logrus.Info("creating initial user")
		hashed, err := bcrypt.GenerateFromPassword([]byte(os.Getenv("DEFAULT_PASSWORD")), bcrypt.DefaultCost)
		if err != nil {
			logrus.Fatal("Error hashing password")
		}
		logrus.Info("Creating default user")
		store.CreateUserNoToken(db, &sidebar.User{
			DisplayName: os.Getenv("DEFAULT_DISPLAYNAME"),
			Email:       os.Getenv("DEFAULT_EMAIL"),
			Password:    hashed,
		})
	}

	server := sidebar.NewServer(auth, create, delete, add, get, up)

	if os.Getenv("PORT") != "" {
		http.ListenAndServe(":"+os.Getenv("PORT"), accessControl(false, server.Serve()))
	} else {
		http.ListenAndServeTLS(":8080", "localhost.pem", "localhost-key.pem", accessControl(true, logging(server.Serve())))
	}
}

// CORS access stuffs
func accessControl(local bool, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if local {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "https://sidebar-frontend.now.sh")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

// log requests
func logging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Printf("%s: %s", r.Method, r.RequestURI)
		h.ServeHTTP(w, r)
	})
}
