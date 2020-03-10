package sidebar

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// type for context.WithValue keys
type ctxKey string

var key []byte

func init() {
	key = []byte("TheKey")
}

// Server returns something that can handle http requests.
type Server interface {
	Serve() *mux.Router
}

type server struct {
	hub    *chathub
	router *mux.Router
	store  *sessions.CookieStore

	// services
	Auth   Authenticater
	Create Creater
	Delete Deleter
	Add    Adder
	Get    Getter
}

// NewServer receives all services needed to provide functionality
// then uses those services to spin-up an HTTP server. A hub for
// handling Websocket connections is also started in a goroutine.
// These things are wrapped in the server and returned.
func NewServer(auth Authenticater, create Creater, delete Deleter, add Adder, get Getter) Server {
	hub := newChathub()

	s := &server{
		hub:    hub,
		store:  sessions.NewCookieStore([]byte("super-secret-key")),
		Auth:   auth,
		Create: create,
		Delete: delete,
		Add:    add,
		Get:    get,
	}

	router := mux.NewRouter().StrictSlash(true)

	// router.Handle("/channels", ).Methods("GET")
	// router.Handle("/sidebars", ).Methods("GET")
	// router.Handle("/messages", ).Methods("GET")
	// router.Handle("/users", ).Methods("GET")

	// router.Handle("/channel/{id}", ).Methods("GET")
	// router.Handle("/sidebar/{id}", ).Methods("GET")
	// router.Handle("/message/{id}", ).Methods("GET")
	// router.Handle("/user/{id}", ).Methods("GET")

	// router.Handle("/channels/", ).Methods("GET")  // r.URL.Query()["user"]
	// router.Handle("/sidebars/", ).Methods("GET")  // r.URL.Query()["user"]
	// router.Handle("/messages/", ).Methods("GET")  // r.URL.Query()["to_user"]
	// router.Handle("/messages/", ).Methods("GET")  // r.URL.Query()["from_user"]
	// router.Handle("/messages/", ).Methods("GET")  // r.URL.Query()["channel"]
	// router.Handle("/users/", ).Methods("GET")  // r.URL.Query()["channel"]
	// router.Handle("/users/", ).Methods("GET")  // r.URL.Query()["sidebar"]

	// router.Handle("/channel", ).Methods("POST")
	// router.Handle("/sidebar", ).Methods("POST")
	// router.Handle("/user", ).Methods("POST")

	// router.Handle("/channel", ).Methods("DELETE")
	// router.Handle("/message", ).Methods("DELETE")
	// router.Handle("/user", ).Methods("DELETE")

	router.Handle("/ws", s.requireAuth(s.HandleWS()))
	router.Handle("/login", s.Login())
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "views/home.html")
	}).Methods("GET")
	s.router = router
	go s.hub.run()

	return s
}

// Return just the mux.Router to be used in http.ListenAndServe.
func (s *server) Serve() *mux.Router {
	return s.router
}

// Login returns an http.HandlerFunc to deal with user attempts to
// log in. The user is authenticated and then a cookie is stored with
// information for later.
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

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"UserID":        user.ID,
			"Email":         user.Email,
			"UserName":      user.DisplayName,
			"Authenticated": true,
			"ExpireAt":      time.Now().Add(time.Minute * 15).Unix(),
		})
		tokenString, err := token.SignedString(key)
		if err != nil {
			http.Error(w, "Unable to sign token", http.StatusInternalServerError)
			return
		}

		session, _ := s.store.Get(r, "chat-cook")
		session.Values["token"] = tokenString
		session.Save(r, w)
	}
}

// requireAuth provides an authentication middleware
func (s *server) requireAuth(f http.HandlerFunc) http.HandlerFunc {

	type Token struct {
		UserID        int
		Email         string
		UserName      string
		Authenticated bool
		ExpireAt      int64
	}

	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.store.Get(r, "chat-cook")

		// Check if user is authenticated
		auth, ok := session.Values["token"].(Token)
		if !ok && auth.Authenticated {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if time.Unix(auth.ExpireAt, 0).Before(time.Now()) {
			http.Error(w, "Expired token", http.StatusForbidden)
			return
		}

		user := User{
			ID:          auth.UserID,
			Email:       auth.Email,
			DisplayName: auth.UserName,
		}

		// refresh the token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"UserID":        user.ID,
			"Email":         user.Email,
			"UserName":      user.DisplayName,
			"Authenticated": true,
			"ExpireAt":      time.Now().Add(time.Minute * 15).Unix(),
		})
		tokenString, err := token.SignedString(key)
		if err != nil {
			http.Error(w, "Unable to sign token", http.StatusInternalServerError)
			return
		}

		session.Values["token"] = tokenString
		session.Save(r, w)

		ctx := context.WithValue(r.Context(), ctxKey("user_info"), user)
		f(w, r.WithContext(ctx))
	}
}

// HandleWS provides a handler for getting Websocket connections setup
// and registering a new client with the hub.
func (s *server) HandleWS() http.HandlerFunc {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.Fatalf("unable to upgrade connection %v", err)
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
