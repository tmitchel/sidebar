package main

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tmitchel/sidebar"
	"github.com/tmitchel/sidebar/mocks/frontend"
	"github.com/tmitchel/sidebar/store"
)

func main() {
	db, err := store.New()
	if err != nil {
		logrus.Fatal(err)
	}
	defer db.Close()

	auth := frontend.NewAuthenticater()
	create := frontend.NewCreater()
	delete := frontend.NewDeleter()
	add := frontend.NewAdder()
	get := frontend.NewGetter()

	server := sidebar.NewServer(auth, create, delete, add, get)
	http.ListenAndServe(":8080", server.Serve())
}
