package sidebar

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type ctxKey string

type server struct {
	hub    *chathub
	router *mux.Router
	store  *sessions.CookieStore
	Auth   Authenticater
	Create Creater
}

func NewServer(auth Authenticater, create Creater) *server {
	hub := NewChathub()

	s := &server{
		hub:    hub,
		store:  sessions.NewCookieStore([]byte("super-secret-key")),
		Auth:   auth,
		Create: create,
	}

	router := mux.NewRouter().StrictSlash(true)
	router.Handle("/ws", s.requireAuth(s.HandleWS()))
	router.Handle("/login", s.Login())
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "views/home.html")
	}).Methods("GET")
	s.router = router
	go s.hub.run()

	return s
}

func (s *server) Serve() *mux.Router {
	return s.router
}

func (s *server) Login() http.HandlerFunc {
	gob.Register(User{})
	type auth struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var auther auth
		if err := json.NewDecoder(r.Body).Decode(&auther); err != nil {
			http.Error(w, "Ill-formatted login attempt", http.StatusBadRequest)
			return
		}

		user, err := s.Auth.Validate(auther.Email, auther.Password)
		if err != nil && user != nil {
			http.Error(w, "Incorrect username/password", http.StatusForbidden)
			return
		}

		session, _ := s.store.Get(r, "chat-cook")
		session.Values["authenticated"] = true
		session.Values["user_info"] = user
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

		var user = User{}
		user, ok := session.Values["user_info"].(User)
		if !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			logrus.Error("Not ok")
			return
		}

		ctx := context.WithValue(r.Context(), ctxKey("user_info"), user)
		f(w, r.WithContext(ctx))
	}
}

func (s *server) HandleWS() http.HandlerFunc {
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

		user, ok := r.Context().Value(ctxKey("user_info")).(User)
		if !ok {
			http.Error(w, "Unable to get user from context", http.StatusInternalServerError)
			return
		}

		cl := &client{
			conn: conn,
			send: make(chan WebSocketMessage),
			hub:  s.hub,
			User: user,
		}

		s.hub.register <- cl

		go cl.writePump()
		go cl.readPump()
	}
}
