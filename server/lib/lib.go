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

package lib

import (
	"github.com/gorilla/mux"

	"./../routes"
)

// Route handler for the server
func Handler() *mux.Router {
	router := mux.NewRouter()
	// GET /client
	// List all registered registration IDs
	router.HandleFunc("/client", routes.ClientIndex).Methods("GET")
	// POST /register
	// Add a new client
	router.HandleFunc("/register", routes.ClientNew).Methods("POST")
	// POST /unregister
	// Remove an existing client
	router.HandleFunc("/unregister", routes.ClientDelete).Methods("POST")

	return router
}
