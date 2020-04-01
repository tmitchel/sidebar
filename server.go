package sidebar

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// type for context.WithValue keys
type ctxKey string

var key []byte

func init() {
	key = []byte("TheKey2")
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
		Auth:   auth,
		Create: create,
		Delete: delete,
		Add:    add,
		Get:    get,
	}

	router := mux.NewRouter().StrictSlash(true)

	apiRouter := router.PathPrefix("/api").Subrouter()

	apiRouter.Handle("/channels", s.requireAuth(s.GetChannels())).Methods("GET")
	apiRouter.Handle("/sidebars", s.requireAuth(s.GetSidebars())).Methods("GET")
	apiRouter.Handle("/messages", s.requireAuth(s.GetMessages())).Methods("GET")
	apiRouter.Handle("/users", s.requireAuth(s.GetUsers())).Methods("GET")

	apiRouter.Handle("/load_channel/{id}", s.requireAuth(s.LoadChannel())).Methods("GET")
	apiRouter.Handle("/load_user/{id}", s.requireAuth(s.LoadUser())).Methods("GET")

	apiRouter.Handle("/channel/{id}", s.requireAuth(s.GetChannel())).Methods("GET")
	apiRouter.Handle("/message/{id}", s.requireAuth(s.GetMessage())).Methods("GET")
	apiRouter.Handle("/user/{id}", s.requireAuth(s.GetUser())).Methods("GET")

	apiRouter.Handle("/channels/", s.requireAuth(s.GetChannelsForUser())).Methods("GET")   // r.URL.Query()["user"]
	apiRouter.Handle("/sidebars/", s.requireAuth(s.GetSidebarsForUser())).Methods("GET")   // r.URL.Query()["user"]
	apiRouter.Handle("/messages/", s.requireAuth(s.GetMessagesToUser())).Methods("GET")    // r.URL.Query()["to_user"]
	apiRouter.Handle("/messages/", s.requireAuth(s.GetMessagesFromUser())).Methods("GET")  // r.URL.Query()["from_user"]
	apiRouter.Handle("/messages/", s.requireAuth(s.GetMessagesInChannel())).Methods("GET") // r.URL.Query()["channel"]
	apiRouter.Handle("/users/", s.requireAuth(s.GetUsersInChannel())).Methods("GET")       // r.URL.Query()["channel"]

	apiRouter.Handle("/channel", s.requireAuth(s.CreateChannel())).Methods("POST")
	apiRouter.Handle("/sidebar/{parent_id}/{user_id}", s.requireAuth(s.CreateSidebar())).Methods("POST")
	apiRouter.Handle("/user", s.CreateUser()).Methods("POST")

	apiRouter.Handle("/add/{user}/{channel}", s.requireAuth(s.AddUserToChannel())).Methods("POST")

	apiRouter.Handle("/channel", s.requireAuth(s.DeleteChannel())).Methods("DELETE")
	apiRouter.Handle("/user", s.requireAuth(s.DeleteUser())).Methods("DELETE")

	apiRouter.Handle("/online_users", s.requireAuth(s.OnlineUsers())).Methods("GET")
	apiRouter.Handle("/refresh_token", s.requireAuth(s.RefreshToken())).Methods("POST")

	router.Handle("/ws", s.requireAuth(s.HandleWS()))
	router.Handle("/login", s.Login()).Methods("POST")
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

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

func logging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Printf("%s: %s", r.Method, r.RequestURI)
		h.ServeHTTP(w, r)
	})
}

func (s *server) LoadChannel() http.HandlerFunc {
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

		users, err := s.Get.GetUsers()
		if err != nil {
			http.Error(w, "Unable to get users for channel", http.StatusInternalServerError)
			return
		}

		messages, err := s.Get.GetMessagesInChannel(reqID)
		if err != nil {
			http.Error(w, "Unable to get messages for channel", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(CompleteChannel{
			Channel:           *channel,
			UsersInChannel:    users,
			MessagesInChannel: messages,
		})
	}
}

func (s *server) LoadUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			http.Error(w, "Unable to convert id", http.StatusBadRequest)
			return
		}
		user, err := s.Get.GetUser(reqID)
		if err != nil {
			http.Error(w, "Unable to get user", http.StatusInternalServerError)
			logrus.Error(err)
			return
		}

		allChannels, err := s.Get.GetChannels()
		if err != nil {
			http.Error(w, "Unable to get all channels", http.StatusInternalServerError)
			logrus.Error(err)
			return
		}

		channels, err := s.Get.GetChannelsForUser(reqID)
		if err != nil {
			http.Error(w, "Unable to get channels for user", http.StatusInternalServerError)
			logrus.Error(err)
			return
		}

		json.NewEncoder(w).Encode(CompleteUser{
			User:            *user,
			ChannelsForUser: channels,
			Channels:        allChannels,
		})
	}
}

func (s *server) OnlineUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users := make([]User, len(s.hub.clients))
		for u := range s.hub.clients {
			users = append(users, u.User)
		}

		json.NewEncoder(w).Encode(users)
	}
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
		channelID, err := strconv.Atoi(r.URL.Query().Get("channel"))
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

func (s *server) CreateSidebar() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var reqChannel Channel
		if err := json.NewDecoder(r.Body).Decode(&reqChannel); err != nil {
			http.Error(w, "Unable to decode new channel", http.StatusBadRequest)
			return
		}

		reqID, err := strconv.Atoi(mux.Vars(r)["parent_id"])
		if err != nil {
			http.Error(w, "Unable to convert id", http.StatusBadRequest)
			return
		}

		reqChannel.IsSidebar = true
		reqChannel.Parent = reqID

		channel, err := s.Create.CreateChannel(&reqChannel)
		if err != nil {
			http.Error(w, "Unable to create sidebar", http.StatusInternalServerError)
			logrus.Error(w, "Error creating sidebar %v", err)
			return
		}

		reqID, err = strconv.Atoi(mux.Vars(r)["user_id"])
		if err != nil {
			http.Error(w, "Unable to convert id", http.StatusBadRequest)
			return
		}

		err = s.Add.AddUserToChannel(reqID, channel.ID)
		if err != nil {
			http.Error(w, "Unable to add user to sidebar", http.StatusInternalServerError)
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

	type Token struct {
		UserID        int
		Email         string
		UserName      string
		Authenticated bool
		jwt.StandardClaims
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("PORT") == "" {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "https://sidebar-frontend.now.sh")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		var auther auth
		if err := json.NewDecoder(r.Body).Decode(&auther); err != nil {
			http.Error(w, "Ill-formatted login attempt", http.StatusBadRequest)
			return
		}

		user, err := s.Auth.Validate(auther.Email, auther.Password)
		if err != nil || user == nil {
			http.Error(w, "Incorrect username/password", http.StatusForbidden)
			return
		}

		expiration := time.Now().Add(time.Minute * 15)
		claims := &Token{
			UserID:        user.ID,
			Email:         user.Email,
			UserName:      user.DisplayName,
			Authenticated: true,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expiration.Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(key)
		if err != nil {
			http.Error(w, "Unable to sign token", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "chat-cook",
			Value:    tokenString,
			Expires:  expiration,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
		})

		json.NewEncoder(w).Encode(user)
	}
}

func (s *server) RefreshToken() http.HandlerFunc {
	type Token struct {
		UserID        int
		Email         string
		UserName      string
		Authenticated bool
		jwt.StandardClaims
	}
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("chat-cook")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			logrus.Errorf("Error with cookie", err)
			return
		}

		tokStr := c.Value
		claims := &Token{}
		tkn, err := jwt.ParseWithClaims(tokStr, claims, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check if user is authenticated
		if !claims.Authenticated {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 90*time.Second {
			w.WriteHeader(http.StatusTooEarly)
			return
		}

		expiration := time.Now().Add(15 * time.Minute)
		claims.ExpiresAt = expiration.Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "chat-cook",
			Value:    tokenString,
			Expires:  expiration,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
		})
	}
}

// requireAuth provides an authentication middleware
func (s *server) requireAuth(f http.HandlerFunc) http.HandlerFunc {

	type Token struct {
		UserID        int
		Email         string
		UserName      string
		Authenticated bool
		jwt.StandardClaims
	}

	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("chat-cook")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			logrus.Errorf("Error with cookie", err)
			return
		}

		tokStr := c.Value
		claims := &Token{}
		tkn, err := jwt.ParseWithClaims(tokStr, claims, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check if user is authenticated
		if !claims.Authenticated {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		user := User{
			ID:          claims.UserID,
			Email:       claims.Email,
			DisplayName: claims.UserName,
		}

		expiration := time.Now().Add(15 * time.Minute)
		claims = &Token{
			UserID:        user.ID,
			Email:         user.Email,
			UserName:      user.DisplayName,
			Authenticated: true,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expiration.Unix(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(key)
		if err != nil {
			http.Error(w, "Unable to sign token", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "chat-cook",
			Value:    tokenString,
			Expires:  expiration,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
		})

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
