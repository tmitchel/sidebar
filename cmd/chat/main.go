package main

import (
	"net/http"

	"github.com/tmitchel/sidebar"
)

func main() {
	server := sidebar.NewServer()
	http.ListenAndServe(":8080", server.Serve())
}
