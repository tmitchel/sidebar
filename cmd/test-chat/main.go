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
