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

	log.Println("Started, serving at 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
