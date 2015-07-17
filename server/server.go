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
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Person is a user who has visited the page.
type Person struct {
	name  string // Name of the person
	count int    // Number of times this person has visited
}

func (p *Person) incrCount() {
	p.count++
}

// Holds the mapping of the person to the person object
var person_req_counts = make(map[string]*Person)

// Handle requests for the homepage.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "/person/:name")
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	person, ok := person_req_counts[name]

	if ok {
		person.incrCount()
		resp := fmt.Sprintf("<h1>%s has visited %v times.</h1>", name, person.count)
		fmt.Fprintln(w, resp)
	} else {
		person_req_counts[name] = &Person{name, 1}
		fmt.Fprintln(w, "<h1>Hello "+name+"</h1>")
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/person/{name}", HelloHandler)

	http.Handle("/", r)

	log.Println("Started, serving at 4260")
	err := http.ListenAndServe(":4260", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
