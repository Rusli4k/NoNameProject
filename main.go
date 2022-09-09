package main

import (
	"log"
	"net/http"
)

func main() {
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/users", createUserHandler)
}

// There is CreateUser endpoint which creates the row in User table.
// If wrong HTTP method is used - response must give understanding about it.+
// If database failed - return ServerError.
// If email already exists - return Conflict.
// If all good - return Success.
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/users" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		userGet()
	case http.MethodPost:
		userCreate()
	case http.MethodPut:
		userUpdate()
	case http.MethodDelete:
		userDelete()
	default:
		http.Error(w, "HTTP method is not allowed", http.StatusMethodNotAllowed)
	}
}

// Serve the resource.
func userGet() {}

// Create a new record.
func userCreate() {}

// Update an existing record.
func userUpdate() {}

// Remove the record.
func userDelete() {}
