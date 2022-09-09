package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

var listOfUsers []userTable

func main() {
	listOfUsers = append(listOfUsers, userTable{
		id:        len(listOfUsers),
		email:     "testEmail@gmail.com",
		fullName:  "Rus Li",
		password:  "qwerty",
		createdAt: time.Now(),
		deleted:   false,
	})

	http.HandleFunc("/users", userHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

type userTable struct {
	id            int
	email         string
	fullName      string
	password      string
	createdAt     time.Time // time of create
	lastUpdatedAt time.Time // time of last update
	deleted       bool      // "true" for deleted users - do not show in "list of users"
}

//type userDB []userTable

// There is CreateUser endpoint which creates the row in User table.
// If wrong HTTP method is used - response must give understanding about it.+
// If database failed - return ServerError.
// If email already exists - return Conflict.
// If all good - return Success.
func userHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/users" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		userGet(w)
	case http.MethodPost:
		userCreate(w)
	case http.MethodPut:
		userUpdate(w)
	case http.MethodDelete:
		userDelete(w)
	default:
		http.Error(w, "HTTP method is not allowed", http.StatusMethodNotAllowed)
	}
}

// Serve the resource.
func userGet(w http.ResponseWriter) {
	w.Write([]byte(fmt.Sprintf("List of users %v ", listOfUsers)))
}

// Create a new record.
func userCreate(w http.ResponseWriter) {
	w.Write([]byte(fmt.Sprintf("Your random number between %v ", "create")))
}

// Update an existing record.
func userUpdate(w http.ResponseWriter) {
	w.Write([]byte(fmt.Sprintf("Your random number between %v ", "update")))
}

// Remove the record.
func userDelete(w http.ResponseWriter) {
	w.Write([]byte(fmt.Sprintf("Your random number between %v ", "delete")))
}
