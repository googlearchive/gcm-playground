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
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

const (
	// What port should the server run on
	port = "4260"

	actionKey         = "action"
	registerNewClient = "register_new_client"
	unregisterClient  = "unregister_client"
	token             = "registration_token"
	stringIdentifier  = "stringIdentifier"
)

var (
	// API key from Cloud console
	// TODO(karangoel): Remove this
	apiKey = "AIzaSyCFVrvWMv0ueY0-wN_RWK_OJ_FmcgkoF_I"

	// GCM sender ID
	// TODO(karangoel): Remove this
	senderId = "1015367374593"

	// The name of the database to connect to
	databaseName = "data.db"

	// Print logging
	debug = true

	// Current database connection
	db gorm.DB
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

// Callback for gcmd listen: check action and dispatch server method
func onMessageReceived(cm gcm.CcsMessage) error {
	log.Printf("Received Message: %+v", cm)

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

		client := Client{token, string_identifier}
		if !ClientExistsInDb(client.RegistrationToken) {
			db.Create(&client)
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
	}
	return nil
}

// Route handler for the server
func Handler() http.Handler {
	router := mux.NewRouter()

	// GET /clients
	// List all registered registration IDs
	router.HandleFunc("/clients", ListClients).Methods("GET")

	// POST /message
	// Send a new message
	router.HandleFunc("/message", SendMessage).Methods("POST")

	return cors.Default().Handler(router)
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

	// Start the server
	log.Println(fmt.Sprintf("Started, serving at port %v", port))
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), Handler())
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}
}
