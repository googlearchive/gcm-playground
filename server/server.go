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
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
	"github.com/google/go-gcm"
)

// What port should the server run on
const port string = "4260"

var (
	// API key from Cloud console
	apiKey = ""

	// GCM sender ID
	senderId = ""

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
	Protocol string `json:"protocol"`
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

// Handle request to save a new client.
// The body of this request must contain a `registration_token`.
// Optionally, the body can contain a `string_identifier` string.
func CreateClient(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))

	if err != nil {
		log.Fatal(err)
	}
	if err := r.Body.Close(); err != nil {
		log.Fatal(err)
	}

	// Decode the passed body into the struct.
	var client Client
	if err := json.Unmarshal(body, &client); err != nil {
		sendUnprocessableEntity(w, err)
	}

	if !ClientExistsInDb(client.RegistrationToken) {
		db.Create(&client)
	}
	w.WriteHeader(http.StatusCreated)
	sendJSON(w, client)
}

// Handle request to delete a client.
// The URL of this request must contain a `registration_token`.
func DeleteClient(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	registration_token := params["registration_token"]

	client := Client{registration_token, ""}

	if !ClientExistsInDb(registration_token) {
		w.WriteHeader(http.StatusNotFound)
	} else {
		db.Delete(&Client{}, "registration_token = ?", client.RegistrationToken)
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handle request to send a new message.
func SendMessage(w http.ResponseWriter, r *http.Request) {
	// TODO(karangoel): Implementation incomplete
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
			log.Println("Message Unmarshal error")
			log.Printf("%+v", err)
			sendUnprocessableEntity(w, err)
			return
		}

		res, sendErr := gcm.SendHttp(apiKey, m)
		if sendErr != nil {
			log.Println("Message send error")
			log.Printf("%+v", sendErr)
			w.WriteHeader(http.StatusInternalServerError)
			sendJSON(w, sendErr)
			return
		}

		log.Printf("%+v", res)
		w.WriteHeader(http.StatusOK)
		sendJSON(w, res)
	} else if protocol == "xmpp" {
		// Send XMPP message
		// TODO(karangoel): Implement this.
		w.WriteHeader(http.StatusOK)
	} else {
		// Error
		w.WriteHeader(http.StatusBadRequest)
		sendJSON(w, &HttpError{"protocol should be HTTP or XMPP only."})
	}
}

// Route handler for the server
func Handler() http.Handler {
	router := mux.NewRouter()

	// GET /clients
	// List all registered registration IDs
	router.HandleFunc("/clients", ListClients).Methods("GET")

	// POST /clients
	// Add a new client
	router.HandleFunc("/clients", CreateClient).Methods("POST")

	// DELETE /clients
	// Remove an existing client
	router.HandleFunc("/clients/{registration_token}", DeleteClient).Methods("DELETE")

	// POST /message
	// Send a new message
	router.HandleFunc("/message", SendMessage).Methods("POST")

	return cors.Default().Handler(router)
}

func main() {
	InitDb()

	// Start the server
	log.Println(fmt.Sprintf("Started, serving at port %v", port))
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), Handler())
	if err != nil {
		log.Fatal("ListenAndServe: " + err.Error())
	}
}
