package sidebar

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type server struct {
	hub    *chathub
	router *mux.Router
	store  *sessions.CookieStore
}

func NewServer() *server {
	hub := NewChathub()

	s := &server{
		hub:   hub,
		store: sessions.NewCookieStore([]byte("super-secret-key")),
	}

	router := mux.NewRouter().StrictSlash(true)
	router.Handle("/ws", s.requireAuth(s.HandleWS()))
	router.Handle("/login", s.Login())
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "views/home.html")
	}).Methods("GET")
	s.router = router

	return s
}

func (s *server) Serve() *mux.Router {
	return s.router
}

func (s *server) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.store.Get(r, "chat-cook")
		session.Values["authenticated"] = true
		session.Save(r, w)
	}
}

func (s *server) requireAuth(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.store.Get(r, "chat-cook")

		// Check if user is authenticated
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		f(w, r)
	}
}

func (s *server) HandleWS() http.HandlerFunc {
	go s.hub.run()
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.Fatalf("unable to upgrade connection", err)
		}

		cl := &client{
			conn: conn,
			send: make(chan WebSocketMessage),
			hub:  s.hub,
		}

		s.hub.register <- cl

		go cl.writePump()
		go cl.readPump()
	}
}
