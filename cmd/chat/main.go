package main

import (
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/services"
	"github.com/tmitchel/sidebar/store"
)

func main() {
	var db store.Database
	if os.Getenv("SBDEV") == "true" {
		logrus.Info("Setting up default database to test frontend")
		var err error
		db, err = store.NewWithMigration("sidebar")
		if err != nil {
			logrus.Fatal(err)
		}

		err = store.MigrationsForTesting(db)
		if err != nil {
			logrus.Error("Error setting up migrations for testing the frontend")
			logrus.Fatal(err)
		}
	} else {
		var err error
		db, err = store.New()
		if err != nil {
			logrus.Fatal(err)
		}
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

	server := sidebar.NewServer(auth, create, delete, add, get)
	http.ListenAndServe(":8080", accessControl(server.Serve()))
}

// CORS access stuffs
func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
