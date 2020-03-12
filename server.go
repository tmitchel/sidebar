package sidebar

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	GetChannel() http.HandlerFunc
	GetMessage() http.HandlerFunc
	GetUser() http.HandlerFunc
	GetSidebars() http.HandlerFunc
	GetChannels() http.HandlerFunc
	GetMessages() http.HandlerFunc
	GetUsers() http.HandlerFunc
	CreateChannel() http.HandlerFunc
	CreateUser() http.HandlerFunc

	// untested...
	AddUserToChannel() http.HandlerFunc
	DeleteChannel() http.HandlerFunc
	DeleteUser() http.HandlerFunc
	Login() http.HandlerFunc
	// requireAuth
	HandleWS() http.HandlerFunc
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

	router.Handle("/channels", s.requireAuth(s.GetChannels())).Methods("GET")
	router.Handle("/sidebars", s.requireAuth(s.GetSidebars())).Methods("GET")
	router.Handle("/messages", s.requireAuth(s.GetMessages())).Methods("GET")
	router.Handle("/users", s.requireAuth(s.GetUsers())).Methods("GET")

	router.Handle("/channel/{id}", s.requireAuth(s.GetChannel())).Methods("GET")
	router.Handle("/message/{id}", s.requireAuth(s.GetMessage())).Methods("GET")
	router.Handle("/user/{id}", s.requireAuth(s.GetUser())).Methods("GET")

	router.Handle("/channels/", s.requireAuth(s.GetChannelsForUser())).Methods("GET")   // r.URL.Query()["user"]
	router.Handle("/sidebars/", s.requireAuth(s.GetSidebarsForUser())).Methods("GET")   // r.URL.Query()["user"]
	router.Handle("/messages/", s.requireAuth(s.GetMessagesToUser())).Methods("GET")    // r.URL.Query()["to_user"]
	router.Handle("/messages/", s.requireAuth(s.GetMessagesFromUser())).Methods("GET")  // r.URL.Query()["from_user"]
	router.Handle("/messages/", s.requireAuth(s.GetMessagesInChannel())).Methods("GET") // r.URL.Query()["channel"]
	router.Handle("/users/", s.requireAuth(s.GetUsersInChannel())).Methods("GET")       // r.URL.Query()["channel"]

	router.Handle("/channel", s.requireAuth(s.CreateChannel())).Methods("POST")
	router.Handle("/user", s.CreateUser()).Methods("POST")

	router.Handle("/add/{user}/{channel}", s.requireAuth(s.AddUserToChannel())).Methods("POST")

	router.Handle("/channel", s.requireAuth(s.DeleteChannel())).Methods("DELETE")
	router.Handle("/user", s.requireAuth(s.DeleteUser())).Methods("DELETE")

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

func logging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Printf("%s: %s", r.Method, r.RequestURI)
		h.ServeHTTP(w, r)
	})
}

func (s *server) GetUsersInChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channelID, err := strconv.Atoi(r.URL.Query().Get("channel"))
		if err != nil {
			http.Error(w, "Error converting channelID", http.StatusBadRequest)
			return
		}

		channels, err := s.Get.GetUsersInChannel(channelID)
		if err != nil {
			http.Error(w, "Error converting channelID", http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(channels)
	}
}

func (s *server) GetChannelsForUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
		if err != nil {
			http.Error(w, "Error converting userID", http.StatusBadRequest)
			return
		}

		channels, err := s.Get.GetChannelsForUser(userID)
		if err != nil {
			http.Error(w, "Error converting userID", http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(channels)
	}
}

func (s *server) GetSidebarsForUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
		if err != nil {
			http.Error(w, "Error converting userID", http.StatusBadRequest)
			return
		}

		channels, err := s.Get.GetChannelsForUser(userID)
		if err != nil {
			http.Error(w, "Error converting userID", http.StatusBadRequest)
			return
		}

		var sidebars []*Channel
		for _, c := range channels {
			if c.IsSidebar {
				sidebars = append(sidebars, c)
			}
		}

		json.NewEncoder(w).Encode(sidebars)
	}
}

func (s *server) GetMessagesToUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(r.URL.Query().Get("to_user"))
		if err != nil {
			http.Error(w, "Error converting userID", http.StatusBadRequest)
			return
		}

		messages, err := s.Get.GetMessagesToUser(userID)
		if err != nil {
			http.Error(w, "Error getting messages", http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(messages)
	}
}

func (s *server) GetMessagesFromUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(r.URL.Query().Get("from_user"))
		if err != nil {
			http.Error(w, "Error converting userID", http.StatusBadRequest)
			return
		}

		messages, err := s.Get.GetMessagesFromUser(userID)
		if err != nil {
			http.Error(w, "Error getting messages", http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(messages)
	}
}

func (s *server) GetMessagesInChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channelID, err := strconv.Atoi(r.URL.Query().Get("channelID"))
		if err != nil {
			http.Error(w, "Error converting channelID", http.StatusBadRequest)
			return
		}

		messages, err := s.Get.GetMessagesInChannel(channelID)
		if err != nil {
			http.Error(w, "Error getting messages", http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(messages)
	}
}

func (s *server) AddUserToChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(mux.Vars(r)["user"])
		if err != nil {
			http.Error(w, "Unable to convert user id", http.StatusBadRequest)
			return
		}

		channelID, err := strconv.Atoi(mux.Vars(r)["channel"])
		if err != nil {
			http.Error(w, "Unable to convert channel id", http.StatusBadRequest)
			return
		}

		if err := s.Add.AddUserToChannel(userID, channelID); err != nil {
			http.Error(w, "Unable to add user to channel", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Successfully added user %v to channel %v", userID, channelID)
	}
}

func (s *server) GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			http.Error(w, "Unable to convert id", http.StatusBadRequest)
			logrus.Errorf("Unable to convert id %v", mux.Vars(r)["id"])
			return
		}

		user, err := s.Get.GetUser(reqID)
		if err != nil {
			http.Error(w, "Unable to get user", http.StatusInternalServerError)
			logrus.Errorf("Unable to get user %v", reqID)
			return
		}

		json.NewEncoder(w).Encode(user)
	}
}

func (s *server) GetChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			http.Error(w, "Unable to convert id", http.StatusBadRequest)
			return
		}
		channel, err := s.Get.GetChannel(reqID)
		if err != nil {
			http.Error(w, "Unable to get channel", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(channel)
	}
}

func (s *server) GetMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			http.Error(w, "Unable to convert id", http.StatusBadRequest)
			return
		}
		message, err := s.Get.GetMessage(reqID)
		if err != nil {
			http.Error(w, "Unable to get message", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(message)
	}
}

func (s *server) GetUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := s.Get.GetUsers()
		if err != nil {
			http.Error(w, "Unable to get users", http.StatusInternalServerError)
			return
		}

		var u []User
		for _, us := range users {
			u = append(u, *us)
		}

		json.NewEncoder(w).Encode(users)
	}
}

func (s *server) GetChannels() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channels, err := s.Get.GetChannels()
		if err != nil {
			http.Error(w, "Unable to get channels", http.StatusInternalServerError)
			logrus.Errorf("Error getting channels %v", err)
			return
		}

		json.NewEncoder(w).Encode(channels)
	}
}

func (s *server) GetSidebars() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channels, err := s.Get.GetChannels()
		if err != nil {
			http.Error(w, "Unable to get channels", http.StatusInternalServerError)
			return
		}

		var sidebars []*Channel
		for _, c := range channels {
			if c.IsSidebar && c.Parent != 0 {
				sidebars = append(sidebars, c)
			}
		}

		json.NewEncoder(w).Encode(sidebars)
	}
}

func (s *server) GetMessages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		messages, err := s.Get.GetMessages()
		if err != nil {
			http.Error(w, "Unable to get messages", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(messages)
	}
}

func (s *server) CreateChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reqChannel Channel
		if err := json.NewDecoder(r.Body).Decode(&reqChannel); err != nil {
			http.Error(w, "Unable to decode new channel", http.StatusBadRequest)
			return
		}

		channel, err := s.Create.CreateChannel(&reqChannel)
		if err != nil {
			http.Error(w, "Unable to create channel", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(channel)
	}
}

func (s *server) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reqUser User
		if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
			http.Error(w, "Unable to decode new user", http.StatusBadRequest)
			return
		}

		user, err := s.Create.CreateUser(&reqUser)
		if err != nil {
			http.Error(w, "Unable to create user", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(user)
	}
}

func (s *server) DeleteChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reqID int
		if err := json.NewDecoder(r.Body).Decode(&reqID); err != nil {
			http.Error(w, "Unable to decode request id", http.StatusBadRequest)
			return
		}

		channel, err := s.Delete.DeleteChannel(reqID)
		if err != nil {
			http.Error(w, "Unable to delete channel", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(channel)
	}
}

func (s *server) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reqID int
		if err := json.NewDecoder(r.Body).Decode(&reqID); err != nil {
			http.Error(w, "Unable to decode request id", http.StatusBadRequest)
			return
		}

		user, err := s.Delete.DeleteUser(reqID)
		if err != nil {
			http.Error(w, "Unable to delete user", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(user)
	}
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
