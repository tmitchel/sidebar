package server

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tmitchel/sidebar"
	"github.com/urfave/negroni"
)

// type for context.WithValue keys
type ctxKey string

type serverError struct {
	Error   error
	Message string
	Status  int
}

// errHandle provides a less verbose way to handle errors
type errHandler func(http.ResponseWriter, *http.Request) *serverError

func (fn errHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		logrus.Errorf("%v", err.Error)
		http.Error(w, err.Message, err.Status)
	}
}

var key []byte

func init() {
	key = []byte("TheKey2")
}

type server struct {
	hub    *chathub
	router *mux.Router

	// services
	Auth   sidebar.Authenticater
	Create sidebar.Creater
	Delete sidebar.Deleter
	Add    sidebar.Adder
	Get    sidebar.Getter
	Up     sidebar.Updater
}

// NewServer receives all services needed to provide functionality
// then uses those services to spin-up an HTTP server. A hub for
// handling Websocket connections is also started in a goroutine.
// These things are wrapped in the server and returned.
func NewServer(auth sidebar.Authenticater, create sidebar.Creater, delete sidebar.Deleter, add sidebar.Adder, get sidebar.Getter, up sidebar.Updater) *server {
	hub := newChathub(create)

	s := &server{
		hub:    hub,
		Auth:   auth,
		Create: create,
		Delete: delete,
		Add:    add,
		Get:    get,
		Up:     up,
	}

	router := mux.NewRouter().StrictSlash(true)
	apiRouter := router.PathPrefix("/api").Subrouter()

	apiRouter.Handle("/channels", s.GetChannels()).Methods("GET")
	apiRouter.Handle("/sidebars", s.GetSidebars()).Methods("GET")
	apiRouter.Handle("/messages", s.GetMessages()).Methods("GET")
	apiRouter.Handle("/users", s.GetUsers()).Methods("GET")

	apiRouter.Handle("/load_channel/{id}", s.LoadChannel()).Methods("GET")
	apiRouter.Handle("/load_user/{id}", s.LoadUser()).Methods("GET")

	apiRouter.Handle("/channel/{id}", s.GetChannel()).Methods("GET")
	apiRouter.Handle("/message/{id}", s.GetMessage()).Methods("GET")
	apiRouter.Handle("/user/{id}", s.GetUser()).Methods("GET")

	apiRouter.Handle("/channels/", s.GetChannelsForUser()).Methods("GET")   // r.URL.Query()["user"]
	apiRouter.Handle("/sidebars/", s.GetSidebarsForUser()).Methods("GET")   // r.URL.Query()["user"]
	apiRouter.Handle("/messages/", s.GetMessagesToUser()).Methods("GET")    // r.URL.Query()["to_user"]
	apiRouter.Handle("/messages/", s.GetMessagesFromUser()).Methods("GET")  // r.URL.Query()["from_user"]
	apiRouter.Handle("/messages/", s.GetMessagesInChannel()).Methods("GET") // r.URL.Query()["channel"]
	apiRouter.Handle("/users/", s.GetUsersInChannel()).Methods("GET")       // r.URL.Query()["channel"]

	apiRouter.Handle("/channel", s.CreateChannel()).Methods("POST")
	apiRouter.Handle("/sidebar/{parent_id}/{user_id}", s.CreateSidebar()).Methods("POST")
	apiRouter.Handle("/direct/{to_id}/{from_id}", s.CreateDirect()).Methods("POST")
	apiRouter.Handle("/user/{create_token}", s.CreateUser()).Methods("POST")
	apiRouter.Handle("/new_token", s.NewToken()).Methods("POST")
	apiRouter.Handle("/resolve/{channel_id}", s.ResolveSidebar()).Methods("POST")
	apiRouter.Handle("/update-userinfo", s.UpdateUserInfo()).Methods("POST")
	apiRouter.Handle("/update-userpass", s.UpdateUserPassword()).Methods("POST")
	apiRouter.Handle("/update-channelinfo", s.UpdateChannelInfo()).Methods("POST")

	apiRouter.Handle("/add/{user}/{channel}", s.AddUserToChannel()).Methods("POST")
	apiRouter.Handle("/leave/{user}/{channel}", s.RemoveUserFromChannel()).Methods("DELETE")

	apiRouter.Handle("/channel", s.DeleteChannel()).Methods("DELETE")
	apiRouter.Handle("/user", s.DeleteUser()).Methods("DELETE")

	apiRouter.Handle("/online_users", s.OnlineUsers()).Methods("GET")
	apiRouter.Handle("/refresh_token", s.RefreshToken()).Methods("POST")
	apiRouter.Use(s.requireAuth) // must be authenticated to use the api endpoints

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
func (s *server) Serve() http.Handler {
	n := negroni.Classic()
	n.UseHandler(s.router)
	return n
}

func (s *server) UpdateUserPassword() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		var payload PasswordUpdate
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			return &serverError{err, "Unable to decode payload", http.StatusBadRequest}
		}

		currentUser, ok := r.Context().Value(ctxKey("user_info")).(sidebar.User)
		if !ok {
			return &serverError{errors.New("Unable to decode user info from context"), "Unable to decode current user", http.StatusBadRequest}
		}

		if payload.ID != currentUser.ID {
			return &serverError{
				errors.Errorf("Request user doesn't match current user. Current: %v Request: %v", currentUser.ID, payload.ID),
				"Request user doesn't match current user.",
				http.StatusBadRequest,
			}
		}

		err := s.Up.UpdateUserPassword(payload.ID, []byte(payload.OldPassword), []byte(payload.NewPassword))
		if err != nil {
			return &serverError{err, "Error updating user info", http.StatusBadRequest}
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Success")
		return nil
	}
}

func (s *server) UpdateUserInfo() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		var reqUser sidebar.User
		if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
			return &serverError{err, "Unable to decode payload", http.StatusBadRequest}
		}

		currentUser, ok := r.Context().Value(ctxKey("user_info")).(sidebar.User)
		if !ok {
			return &serverError{errors.New("Unable to decode user info from context"), "Unable to decode current user", http.StatusBadRequest}
		}

		if reqUser.ID != currentUser.ID {
			return &serverError{
				errors.Errorf("Request user doesn't match current user. Current: %v Request: %v", currentUser.ID, reqUser.ID),
				"Request user doesn't match current user.",
				http.StatusBadRequest,
			}
		}

		err := s.Up.UpdateUserInfo(&reqUser)
		if err != nil {
			return &serverError{err, "Error updating user info", http.StatusBadRequest}
		}

		newUser, err := s.Get.GetUser(reqUser.ID)
		if err != nil {
			return &serverError{err, "Error getting updated user", http.StatusBadRequest}
		}

		json.NewEncoder(w).Encode(newUser)
		return nil
	}
}

func (s *server) UpdateChannelInfo() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		var reqChannel sidebar.Channel
		if err := json.NewDecoder(r.Body).Decode(&reqChannel); err != nil {
			return &serverError{err, "Unable to decode payload", http.StatusBadRequest}
		}

		currentUser, ok := r.Context().Value(ctxKey("user_info")).(sidebar.User)
		if !ok {
			return &serverError{errors.New("Unable to decode user info from context"), "Unable to decode current user", http.StatusBadRequest}
		}

		members, err := s.Get.GetUsersInChannel(reqChannel.ID)
		if err != nil {
			return &serverError{err, "Unable to get memebers of the channel", http.StatusBadRequest}
		}

		var found bool
		for _, m := range members {
			if m.ID == currentUser.ID {
				found = true
				break
			}
		}

		if !found {
			return &serverError{err, "Cannot update channel that you aren't a part of", http.StatusBadRequest}
		}

		err = s.Up.UpdateChannelInfo(&reqChannel)
		if err != nil {
			return &serverError{err, "Error updating channel info", http.StatusBadRequest}
		}

		newChannel, err := s.Get.GetChannel(reqChannel.ID)
		if err != nil {
			return &serverError{err, "Error getting updated channel", http.StatusBadRequest}
		}

		json.NewEncoder(w).Encode(newChannel)
		return nil
	}
}

func (s *server) LoadChannel() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		reqID := mux.Vars(r)["id"]
		channel, err := s.Get.GetChannel(reqID)
		if err != nil {
			return &serverError{err, "Unable to get channel id from request param", http.StatusInternalServerError}
		}

		users, err := s.Get.GetUsers()
		if err != nil {
			return &serverError{err, "Unable to get users for channel", http.StatusInternalServerError}
		}

		messages, err := s.Get.GetMessagesInChannel(reqID)
		if err != nil {
			return &serverError{err, "Unable to get messages for channel", http.StatusInternalServerError}
		}

		json.NewEncoder(w).Encode(ChannelWithUsersAndMessages{
			Channel:           *channel,
			UsersInChannel:    users,
			MessagesInChannel: messages,
		})
		return nil
	}
}

func (s *server) LoadUser() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		reqID := mux.Vars(r)["id"]
		user, err := s.Get.GetUser(reqID)
		if err != nil {
			return &serverError{err, "Unable to get user id from request param", http.StatusInternalServerError}
		}

		allChannels, err := s.Get.GetChannels()
		if err != nil {
			return &serverError{err, "Unable to get all channels", http.StatusInternalServerError}
		}

		channelsForUser, err := s.Get.GetChannelsForUser(reqID)
		if err != nil {
			return &serverError{err, "Unable to get channels for user", http.StatusInternalServerError}
		}

		channelWithInfo := make([]*ChannelWithMemberInfo, len(allChannels))
		var matched bool
		for i, c := range allChannels {
			matched = false
			for _, cc := range channelsForUser {
				if c.ID == cc.ID {
					matched = true
					break
				}
			}
			channelWithInfo[i] = &ChannelWithMemberInfo{Channel: *c, Member: matched}
		}

		json.NewEncoder(w).Encode(UserWithChannels{
			User:     *user,
			Channels: channelWithInfo,
		})
		return nil
	}
}

func (s *server) OnlineUsers() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		users := make([]sidebar.User, len(s.hub.clients))
		for u := range s.hub.clients {
			users = append(users, u.User)
		}

		json.NewEncoder(w).Encode(users)
		return nil
	}
}

func (s *server) GetUsersInChannel() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		channelID := r.URL.Query().Get("channel")
		users, err := s.Get.GetUsersInChannel(channelID)
		if err != nil {
			return &serverError{err, "Error getting users in the channel", http.StatusBadRequest}
		}

		json.NewEncoder(w).Encode(users)
		return nil
	}
}

func (s *server) GetChannelsForUser() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		userID := r.URL.Query().Get("user_id")
		channels, err := s.Get.GetChannelsForUser(userID)
		if err != nil {
			return &serverError{err, "Error getting channels for the user", http.StatusBadRequest}
		}

		json.NewEncoder(w).Encode(channels)
		return nil
	}
}

func (s *server) GetSidebarsForUser() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		userID := r.URL.Query().Get("user_id")
		channels, err := s.Get.GetChannelsForUser(userID)
		if err != nil {
			return &serverError{err, "Error getting channels for the user", http.StatusBadRequest}
		}

		var sidebars []*sidebar.Channel
		for _, c := range channels {
			if c.IsSidebar {
				sidebars = append(sidebars, c)
			}
		}

		json.NewEncoder(w).Encode(sidebars)
		return nil
	}
}

func (s *server) GetMessagesToUser() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		userID := r.URL.Query().Get("to_user")
		messages, err := s.Get.GetMessagesToUser(userID)
		if err != nil {
			return &serverError{err, "Error getting messages to the user", http.StatusBadRequest}
		}

		json.NewEncoder(w).Encode(messages)
		return nil
	}
}

func (s *server) GetMessagesFromUser() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		userID := r.URL.Query().Get("from_user")
		messages, err := s.Get.GetMessagesFromUser(userID)
		if err != nil {
			return &serverError{err, "Error getting messages from the user", http.StatusBadRequest}
		}

		json.NewEncoder(w).Encode(messages)
		return nil
	}
}

func (s *server) GetMessagesInChannel() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		channelID := r.URL.Query().Get("channel")
		messages, err := s.Get.GetMessagesInChannel(channelID)
		if err != nil {
			return &serverError{err, "Error getting messages in the channel", http.StatusBadRequest}
		}

		json.NewEncoder(w).Encode(messages)
		return nil
	}
}

func (s *server) AddUserToChannel() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		userID := mux.Vars(r)["user"]
		channelID := mux.Vars(r)["channel"]
		if err := s.Add.AddUserToChannel(userID, channelID); err != nil {
			return &serverError{err, "Unable to add user to channel", http.StatusInternalServerError}
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Successfully added user %v to channel %v", userID, channelID)
		return nil
	}
}

func (s *server) RemoveUserFromChannel() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		userID := mux.Vars(r)["user"]
		channelID := mux.Vars(r)["channel"]
		if err := s.Add.RemoveUserFromChannel(userID, channelID); err != nil {
			return &serverError{err, "Unable to remove user from channel", http.StatusInternalServerError}
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Successfully removed user %v from channel %v", userID, channelID)
		return nil
	}
}

func (s *server) GetUser() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		reqID := mux.Vars(r)["id"]
		user, err := s.Get.GetUser(reqID)
		if err != nil {
			return &serverError{err, "Unable to get user id from request param", http.StatusInternalServerError}
		}

		json.NewEncoder(w).Encode(user)
		return nil
	}
}

func (s *server) GetChannel() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		reqID := mux.Vars(r)["id"]
		channel, err := s.Get.GetChannel(reqID)
		if err != nil {
			return &serverError{err, "Unable to get channel id from request param", http.StatusInternalServerError}
		}

		json.NewEncoder(w).Encode(channel)
		return nil
	}
}

func (s *server) GetMessage() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		reqID := mux.Vars(r)["id"]
		message, err := s.Get.GetMessage(reqID)
		if err != nil {
			return &serverError{err, "Unable to get message id from request param", http.StatusInternalServerError}
		}

		json.NewEncoder(w).Encode(message)
		return nil
	}
}

func (s *server) GetUsers() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		users, err := s.Get.GetUsers()
		if err != nil {
			return &serverError{err, "Unable to get users", http.StatusInternalServerError}
		}

		var u []sidebar.User
		for _, us := range users {
			u = append(u, *us)
		}

		json.NewEncoder(w).Encode(users)
		return nil
	}
}

func (s *server) GetChannels() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		channels, err := s.Get.GetChannels()
		if err != nil {
			return &serverError{err, "Unable to get channels", http.StatusInternalServerError}
		}

		json.NewEncoder(w).Encode(channels)
		return nil
	}
}

func (s *server) GetSidebars() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		channels, err := s.Get.GetChannels()
		if err != nil {
			return &serverError{err, "Unable to get sidebars", http.StatusInternalServerError}
		}

		var sidebars []*sidebar.Channel
		for _, c := range channels {
			if c.IsSidebar && c.Parent != "" {
				sidebars = append(sidebars, c)
			}
		}

		json.NewEncoder(w).Encode(sidebars)
		return nil
	}
}

func (s *server) GetMessages() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		messages, err := s.Get.GetMessages()
		if err != nil {
			return &serverError{err, "Unable to get messages", http.StatusInternalServerError}
		}

		json.NewEncoder(w).Encode(messages)
		return nil
	}
}

func (s *server) CreateChannel() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		var reqChannel sidebar.Channel
		if err := json.NewDecoder(r.Body).Decode(&reqChannel); err != nil {
			return &serverError{err, "Unable to decode payload", http.StatusBadRequest}
		}

		channel, err := s.Create.CreateChannel(&reqChannel)
		if err != nil {
			return &serverError{err, "Unable to create channel", http.StatusInternalServerError}
		}

		json.NewEncoder(w).Encode(channel)
		return nil
	}
}

func (s *server) CreateDirect() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		var reqChannel sidebar.Channel
		if err := json.NewDecoder(r.Body).Decode(&reqChannel); err != nil {
			return &serverError{err, "Unable to decode payload", http.StatusBadRequest}
		}

		toID := mux.Vars(r)["to_id"]
		fromID := mux.Vars(r)["from_id"]
		reqChannel.Direct = true
		channel, err := s.Create.CreateChannel(&reqChannel)
		if err != nil {
			return &serverError{err, "Unable to create direct channel", http.StatusInternalServerError}
		}

		err = s.Add.AddUserToChannel(toID, channel.ID)
		if err != nil {
			return &serverError{err, "Unable to add 'to' user to channel", http.StatusInternalServerError}
		}

		err = s.Add.AddUserToChannel(fromID, channel.ID)
		if err != nil {
			return &serverError{err, "Unable to add 'from' user to channel", http.StatusInternalServerError}
		}

		json.NewEncoder(w).Encode(channel)
		return nil
	}
}

func (s *server) CreateSidebar() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		var reqChannel sidebar.Channel
		if err := json.NewDecoder(r.Body).Decode(&reqChannel); err != nil {
			return &serverError{err, "Unable to decode payload", http.StatusBadRequest}
		}

		reqChannel.IsSidebar = true
		reqChannel.Parent = mux.Vars(r)["parent_id"]

		channel, err := s.Create.CreateChannel(&reqChannel)
		if err != nil {
			return &serverError{err, "Unable to create sidebar", http.StatusInternalServerError}
		}

		members, err := s.Get.GetUsersInChannel(reqChannel.Parent)
		if err != nil {
			return &serverError{err, "Unable to get users from parent channel", http.StatusInternalServerError}
		}

		for _, member := range members {
			err = s.Add.AddUserToChannel(member.ID, channel.ID)
			if err != nil {
				return &serverError{err, "Unable to add user to sidebar", http.StatusInternalServerError}
			}
		}

		json.NewEncoder(w).Encode(channel)
		return nil
	}
}

func (s *server) CreateUser() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		token := mux.Vars(r)["create_token"]
		var reqUser SignupUser
		if err := json.NewDecoder(r.Body).Decode(&reqUser); err != nil {
			return &serverError{err, "Unable to decode payload", http.StatusBadRequest}
		}

		converted := sidebar.User{
			ID:          reqUser.ID,
			DisplayName: reqUser.DisplayName,
			Email:       reqUser.Email,
			Password:    []byte(reqUser.Password),
			ProfileImg:  reqUser.ProfileImg,
		}
		user, err := s.Create.CreateUser(&converted, token)
		if err != nil {
			return &serverError{err, "Unable to create user", http.StatusInternalServerError}
		}

		expiration := time.Now().Add(time.Minute * 15)
		claims := &JWTToken{
			UserID:        user.ID,
			Email:         user.Email,
			UserName:      user.DisplayName,
			Authenticated: true,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expiration.Unix(),
			},
		}

		userToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := userToken.SignedString(key)
		if err != nil {
			return &serverError{err, "Unable to sign token", http.StatusInternalServerError}
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
		return nil
	}
}

func (s *server) NewToken() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		user, ok := r.Context().Value(ctxKey("user_info")).(sidebar.User)
		if !ok {
			return &serverError{errors.New("Unable to decode user info from context"), "Unable to decode current user", http.StatusBadRequest}
		}

		token, err := s.Create.NewToken(user.ID)
		if err != nil {
			return &serverError{err, "Error creating token", http.StatusInternalServerError}
		}

		json.NewEncoder(w).Encode(struct{ Token string }{token})
		return nil
	}
}

func (s *server) ResolveSidebar() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		sid := mux.Vars(r)["channel_id"]
		err := s.Add.ResolveChannel(sid)
		if err != nil {
			return &serverError{err, "Unable to resolve channel", http.StatusInternalServerError}
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Success")
		return nil
	}
}

func (s *server) DeleteChannel() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		var reqID string
		if err := json.NewDecoder(r.Body).Decode(&reqID); err != nil || reqID == "" {
			return &serverError{err, "Unable to decode payload", http.StatusBadRequest}
		}

		channel, err := s.Delete.DeleteChannel(reqID)
		if err != nil {
			return &serverError{err, "Unable to delete channel", http.StatusInternalServerError}
		}

		json.NewEncoder(w).Encode(channel)
		return nil
	}
}

func (s *server) DeleteUser() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		var reqID string
		if err := json.NewDecoder(r.Body).Decode(&reqID); err != nil || reqID == "" {
			return &serverError{err, "Unable to decode payload", http.StatusBadRequest}
		}

		user, err := s.Delete.DeleteUser(reqID)
		if err != nil {
			return &serverError{err, "Unable to delete user", http.StatusInternalServerError}
		}

		json.NewEncoder(w).Encode(user)
		return nil
	}
}

// Login returns an errHandler to deal with user attempts to
// log in. The user is authenticated and then a cookie is stored with
// information for later.
func (s *server) Login() errHandler {
	gob.Register(sidebar.User{})
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		if os.Getenv("PORT") == "" {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "https://sidebar-frontend.now.sh")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		var auther AuthInfo
		if err := json.NewDecoder(r.Body).Decode(&auther); err != nil {
			return &serverError{err, "Ill-formatted login attempt", http.StatusBadRequest}
		}

		user, err := s.Auth.Validate(auther.Email, auther.Password)
		if err != nil || user == nil {
			return &serverError{err, "Incorrect username/password", http.StatusForbidden}
		}

		expiration := time.Now().Add(time.Minute * 15)
		claims := &JWTToken{
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
			return &serverError{err, "Unable to sign token", http.StatusInternalServerError}
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
		return nil
	}
}

func (s *server) RefreshToken() errHandler {
	return func(w http.ResponseWriter, r *http.Request) *serverError {
		c, err := r.Cookie("chat-cook")
		if err != nil {
			return &serverError{err, "Error with cookie", http.StatusUnauthorized}
		}

		tokStr := c.Value
		claims := &JWTToken{}
		tkn, err := jwt.ParseWithClaims(tokStr, claims, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})

		if err != nil {
			return &serverError{err, "Error parsing JWT", http.StatusUnauthorized}
		} else if !tkn.Valid {
			return &serverError{err, "Error token is not valid", http.StatusUnauthorized}
		}

		// Check if user is authenticated
		if !claims.Authenticated {
			return &serverError{err, "Error user not authenticated", http.StatusUnauthorized}
		}

		if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 90*time.Second {
			return &serverError{nil, "Not ready", http.StatusTooEarly}
		}

		expiration := time.Now().Add(15 * time.Minute)
		claims.ExpiresAt = expiration.Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(key)
		if err != nil {
			return &serverError{err, "Error signing token", http.StatusInternalServerError}
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "chat-cook",
			Value:    tokenString,
			Expires:  expiration,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
		})
		return nil
	}
}

// requireAuth provides an authentication middleware
func (s *server) requireAuth(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("chat-cook")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			logrus.Error("Error with cookie", err)
			return
		}

		tokStr := c.Value
		claims := &JWTToken{}
		tkn, err := jwt.ParseWithClaims(tokStr, claims, func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logrus.Error(err)
			return
		} else if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			logrus.Error(err)
			return
		}

		// Check if user is authenticated
		if !claims.Authenticated {
			http.Error(w, "Forbidden", http.StatusForbidden)
			logrus.Error(err)
			return
		}

		user := sidebar.User{
			ID:          claims.UserID,
			Email:       claims.Email,
			DisplayName: claims.UserName,
		}

		ctx := context.WithValue(r.Context(), ctxKey("user_info"), user)
		f.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HandleWS provides a handler for getting Websocket connections setup
// and registering a new client with the hub.
func (s *server) HandleWS() errHandler {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return func(w http.ResponseWriter, r *http.Request) *serverError {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.Fatalf("unable to upgrade connection %v", err)
		}

		user, ok := r.Context().Value(ctxKey("user_info")).(sidebar.User)
		if !ok {
			return &serverError{errors.New("Unable to decode user info from context"), "Unable to decode current user", http.StatusBadRequest}
		}

		cl := &client{
			conn: conn,
			send: make(chan sidebar.WebSocketMessage),
			hub:  s.hub,
			User: user,
		}

		s.hub.register <- cl

		go cl.writePump()
		go cl.readPump()
		return nil
	}
}
