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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	server        *httptest.Server
	clientUrl     string
	registerUrl   string
	unregisterUrl string
)

func init() {
	databaseName = "data-test.db"
	debug = false
	InitDb()
	server = httptest.NewServer(Handler())
	clientUrl = fmt.Sprintf("%s/clients", server.URL)
	registerUrl = fmt.Sprintf("%s/clients", server.URL)
	unregisterUrl = fmt.Sprintf("%s/clients", server.URL)
}

func teardownTest() {
	db.Exec("DELETE FROM clients;")
}

func assertEqual(t *testing.T, v, e interface{}) {
	if v != e {
		t.Fatalf("%#v != %#v", v, e)
	}
}

// Make an HTTP request
func MakeRequest(method string, url string, body string) (resp *http.Response, err error) {
	request, _ := http.NewRequest(method, url, strings.NewReader(body))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "applicaiton/json")
	return http.DefaultClient.Do(request)
}

// Makes a request with the passed data to register a new client
func RegisterClient(body string) (resp *http.Response, err error) {
	return MakeRequest("POST", registerUrl, body)
}

// Makes a request with the passed data to unregister a new client
func UnregisterClient(reg_id string) (resp *http.Response, err error) {
	return MakeRequest("DELETE", fmt.Sprintf("%s/%s", unregisterUrl, reg_id), "")
}

// Test getting a list of all clients
func TestListClients(t *testing.T) {
	defer teardownTest()

	request, err := http.NewRequest("GET", clientUrl, nil)
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Error(err)
	}
	assertEqual(t, res.StatusCode, http.StatusOK)
}

// Test registering a client
func TestRegisterClient(t *testing.T) {
	defer teardownTest()

	clientJson := `{"registration_token": "Est deserunt eu elit in."}`
	res, err := RegisterClient(clientJson)
	if err != nil {
		t.Error(err)
	}
	assertEqual(t, res.StatusCode, http.StatusCreated)
}

// Test registering a client multiple times
func TestRegisterClientMultipleTimes(t *testing.T) {
	defer teardownTest()

	clientJson := `{"registration_token": "Est deserunt eu elit in."}`
	res, err := RegisterClient(clientJson)
	if err != nil {
		t.Error(err)
	}
	res, err = RegisterClient(clientJson)
	if err != nil {
		t.Error(err)
	}
	res, err = RegisterClient(clientJson)
	if err != nil {
		t.Error(err)
	}
	assertEqual(t, res.StatusCode, http.StatusCreated)
}

// Test registering multiple clients
func TestRegisterMultipleClients(t *testing.T) {
	defer teardownTest()

	testData := []string{
		`{"registration_token": "900150983cd24fb0d6963f7d28e17f72", "string_identifier": "Lorem ipsum Cillum in anim culpa labore."}`,
		`{"registration_token": "f0e57d481af6ac8aaad01a78eaa394d9"}`,
		`{"registration_token": "0be4a67ee30268d79bfb3709702ec59c", "string_identifier": "Lorem ipsum Esse nostrud irure laborum incididunt sit."}`,
		`{"registration_token": "bc5a1138d76b4aa49ea6f826320dc6e5"}`,
	}

	for _, clientJson := range testData {
		res, err := RegisterClient(clientJson)
		if err != nil {
			t.Error(err)
		}
		assertEqual(t, res.StatusCode, http.StatusCreated)
	}
}

// Test unregistering a client
func TestUnregisterClient(t *testing.T) {
	defer teardownTest()

	reg_id := "Dolor qui occaecat proident."
	clientJson := `{"registration_token": "Dolor qui occaecat proident."}`
	RegisterClient(clientJson)
	res, err := UnregisterClient(reg_id)
	if err != nil {
		t.Error(err)
	}
	assertEqual(t, res.StatusCode, http.StatusNoContent)
}

// Test unregistering a client that's not registered
func TestUnregisterClientNotRegistered(t *testing.T) {
	defer teardownTest()

	reg_id := "Dolor qui occaecat proident."
	res, err := UnregisterClient(reg_id)
	if err != nil {
		t.Error(err)
	}
	assertEqual(t, res.StatusCode, http.StatusNotFound)
}

// Test unregistering multiple clients
func TestUnregisterMultipleClients(t *testing.T) {
	defer teardownTest()

	testData := []string{
		`{"registration_token": "900150983cd24fb0d6963f7d28e17f72"}`,
		`{"registration_token": "f0e57d481af6ac8aaad01a78eaa394d9"}`,
		`{"registration_token": "0be4a67ee30268d79bfb3709702ec59c"}`,
		`{"registration_token": "bc5a1138d76b4aa49ea6f826320dc6e5"}`,
	}

	testRegIds := []string{
		"900150983cd24fb0d6963f7d28e17f72",
		"f0e57d481af6ac8aaad01a78eaa394d9",
		"0be4a67ee30268d79bfb3709702ec59c",
		"bc5a1138d76b4aa49ea6f826320dc6e5",
	}

	for _, reg_id := range testData {
		RegisterClient(reg_id)
	}

	for _, reg_id := range testRegIds {
		res, err := UnregisterClient(reg_id)
		if err != nil {
			t.Error(err)
		}
		assertEqual(t, res.StatusCode, http.StatusNoContent)
	}
}
