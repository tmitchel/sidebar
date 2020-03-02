package main

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/services"
	"github.com/tmitchel/sidebar/store"
)

func main() {
	db, err := store.New()
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

	server := sidebar.NewServer(auth, create)
	http.ListenAndServe(":8080", server.Serve())
}
