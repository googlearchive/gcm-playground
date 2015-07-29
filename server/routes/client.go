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

package routes

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"./../models"
)

// Checks if the passed registration_id exists in the database
func ExistsInDb(registration_id string) bool {
	count := 0
	db.Model(models.Client{}).Where("registration_id = ?", registration_id).Count(&count)
	return count != 0
}

// Handle requests to get all the clients in the database.
func ClientIndex(w http.ResponseWriter, r *http.Request) {
	var clients []models.Client
	if err := db.Find(&clients).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	clientArray := models.ClientCollection{Data: clients}
	sendJSON(w, clientArray)
}

// Handle request to save a new client.
// The body of this request must contain a `registration_id`.
// Optionally, the body can contain a `data` string.
func ClientNew(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))

	if err != nil {
		log.Fatal(err)
	}
	if err := r.Body.Close(); err != nil {
		log.Fatal(err)
	}

	// Decode the passed body into the struct.
	var client models.Client
	if err := json.Unmarshal(body, &client); err != nil {
		sendUnprocessableEntity(w, err)
	}

	if !ExistsInDb(client.RegistrationId) {
		db.Create(&client)
		// log.Printf("Client #%d created", client.RegistrationId)
	}
	w.WriteHeader(http.StatusCreated)
	sendJSON(w, client)
}

// Handle request to delete a client.
// The body of this request must contain a `registration_id`.
func ClientDelete(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Fatal(err)
	}
	if err := r.Body.Close(); err != nil {
		log.Fatal(err)
	}

	var client models.Client
	if err := json.Unmarshal(body, &client); err != nil {
		sendUnprocessableEntity(w, err)
	}

	if !ExistsInDb(client.RegistrationId) {
		w.WriteHeader(http.StatusNotFound)
	} else {
		db.Delete(&models.Client{}, "registration_id = ?", client.RegistrationId)
		w.WriteHeader(http.StatusNoContent)
	}
}
