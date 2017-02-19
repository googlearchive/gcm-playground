// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/google/go-gcm"
	"github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

const (
	// What port should the server run on
	port = "4260"

	actionKey          = "action"
	registerNewClient  = "register_new_client"
	unregisterClient   = "unregister_client"
	token              = "registration_token"
	stringIdentifier   = "stringIdentifier"
	statusRegistered   = "registered"
	statusUnregistered = "unregistered"
)

var (
	// API key from Cloud console
	apiKey = "AIzaSyDc180dAAdcfZT77cuyx-BzIGgnmINivqI"

	// GCM sender ID
	senderId = "1099444446302"

	// The name of the database to connect to
	databaseName = "data.db"

	// Print logging
	debug = true

	// Current database connection
	db *gorm.DB

	// Websocket connection
	socket socketio.Socket
)

type Client struct {
	RegistrationToken string `sql:"not null;unique" json:"registration_token" gorm:"primary_key"`
	StringIdentifier  string `json:"string_identifier"`
}

type ClientCollection struct {
	Clients []Client `json:"clients"`
}

type DownstreamMessage struct {
	Protocol string          `json:"protocol"`
	Message  json.RawMessage `json:"message"`
}

type HttpError struct {
	Error string `json:"error"`
}

// Checks if the passed registration_token exists in the database
func ClientExistsInDb(RegistrationToken string) bool {
	count := 0
	db.Model(Client{}).Where("registration_token = ?", RegistrationToken).Count(&count)
	return count != 0
}

func InitDb() {
	// Database connection
	var err error
	db, err = gorm.Open("sqlite3", databaseName)
	if err != nil {
		log.Fatal(err)
	}
	db.DB()
	db.AutoMigrate(&Client{})
	db.LogMode(debug) // Helps with debugging
}

func sendJSON(w http.ResponseWriter, obj interface{}) {
	json.NewEncoder(w).Encode(obj)
}

func sendUnprocessableEntity(w http.ResponseWriter, err error) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusNotAcceptable)
	return json.NewEncoder(w).Encode(err)
}

// Handle requests to get all the clients in the database.
func ListClients(w http.ResponseWriter, r *http.Request) {
	var clients []Client
	if err := db.Find(&clients).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	clientArray := ClientCollection{Clients: clients}
	sendJSON(w, clientArray)
}

func SendOkResponse(w http.ResponseWriter, res interface{}) {
	log.Printf("Response: %+v", res)
	w.WriteHeader(http.StatusOK)
	sendJSON(w, res)
}

func SendMessageSendError(w http.ResponseWriter, sendErr error) {
	log.Println("Message send error: %+v", sendErr)
	w.WriteHeader(http.StatusInternalServerError)
	sendJSON(w, sendErr)
}

// Handle request to send a new message.
func SendMessage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))

	if err != nil {
		log.Fatal(err)
	}
	if err := r.Body.Close(); err != nil {
		log.Fatal(err)
	}

	// Decode the passed body into the struct.
	var message DownstreamMessage
	if err := json.Unmarshal(body, &message); err != nil {
		sendUnprocessableEntity(w, err)
		return
	}

	protocol := strings.ToLower(message.Protocol)

	if protocol == "http" {
		// Send HTTP message
		var m gcm.HttpMessage
		if err := json.Unmarshal(message.Message, &m); err != nil {
			log.Println("Message Unmarshal error: %+v", err)
			sendUnprocessableEntity(w, err)
			return
		}

		res, sendErr := gcm.SendHttp(apiKey, m)
		if sendErr != nil {
			SendMessageSendError(w, sendErr)
		} else {
			SendOkResponse(w, res)
		}
	} else if protocol == "xmpp" {
		// Send XMPP message
		var m gcm.XmppMessage
		if err := json.Unmarshal(message.Message, &m); err != nil {
			log.Println("Message Unmarshal error: %+v", err)
			sendUnprocessableEntity(w, err)
			return
		}

		res, _, sendErr := gcm.SendXmpp(senderId, apiKey, m)
		if sendErr != nil {
			SendMessageSendError(w, sendErr)
		} else {
			SendOkResponse(w, res)
		}
	} else {
		// Error
		w.WriteHeader(http.StatusBadRequest)
		sendJSON(w, &HttpError{"protocol should be HTTP or XMPP only."})
	}
}

func SendClientStatus(token string, d gcm.Data) error {
	m := gcm.XmppMessage{
		To:       token,
		Priority: "high",
		Data:     d,
	}
	_, _, sendErr := gcm.SendXmpp(senderId, apiKey, m)
	if sendErr != nil {
		return fmt.Errorf("sending ack failed: %v", sendErr)
	}
	return nil
}

// Callback for gcmd listen: check action and dispatch server method
func onMessageReceived(cm gcm.CcsMessage) error {
	log.Printf("Received Message: %+v", cm)

	if socket != nil {
		log.Println("emit:", socket.Emit("upstream message", cm))
	}

	d := cm.Data

	switch d[actionKey] {
	case registerNewClient:
		token, ok := d[token].(string)
		if !ok {
			return errors.New("Error decoding registration token for new client.")
		}
		string_identifier, ok := d[stringIdentifier].(string)
		if !ok {
			return errors.New("Error decoding string identifier for new client.")
		}

		if !ClientExistsInDb(token) {
			client := Client{token, string_identifier}
			db.Create(&client)
		} else {
			db.Model(Client{}).Where("registration_token = ?", token).Update("string_identifier", string_identifier)
		}

		// Send the client registered status.
		err := SendClientStatus(token, gcm.Data{actionKey: registerNewClient, "status": statusRegistered})
		if err != nil {
			log.Println(err)
		}
	case unregisterClient:
		token, ok := d[token].(string)
		if !ok {
			return errors.New("Error decoding registration token for client.")
		}

		client := Client{token, ""}

		if !ClientExistsInDb(token) {
			return errors.New("Client does not exist in database.")
		} else {
			db.Delete(&Client{}, "registration_token = ?", client.RegistrationToken)
		}

		// Send the client registered status.
		err := SendClientStatus(token, gcm.Data{actionKey: unregisterClient, "status": statusUnregistered})
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

// Route handler for the server
func Handler() http.Handler {
	router := mux.NewRouter()

	// Set up the websocket server
	wsServer, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	wsServer.On("connection", func(so socketio.Socket) {
		log.Println("on connection")
		socket = so
	})
	router.Handle("/socket.io/", wsServer)

	// GET /clients
	// List all registered registration IDs
	router.HandleFunc("/clients", ListClients).Methods("GET")

	// POST /message
	// Send a new message
	router.HandleFunc("/message", SendMessage).Methods("POST")

	c := cors.New(cors.Options{
		AllowCredentials: true,
	})
	return c.Handler(router)
}

func main() {
	InitDb()

	gcm.DebugMode = true
	go func() {
		err := gcm.Listen(senderId, apiKey, onMessageReceived, nil)
		if err != nil {
			fmt.Printf("Listen error: %v", err)
		}
	}()

	//Start the server
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), Handler())
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}
	log.Println(fmt.Sprintf("Started, serving at port %v", port))
}
