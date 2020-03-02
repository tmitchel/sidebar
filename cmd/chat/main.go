package main

import (
	"net/http"

	"github.com/tmitchel/chat"
)

func main() {
	server := chat.NewServer()
	http.ListenAndServe(":8080", server.Serve())
}
