package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var user User
var users []User
var idCount int

var (
	errInvalidEmail    = errors.New(`{"error":"incorrect email input","details":"email must have 5-256 chars and contain "@""}`)
	errConflictEmail   = errors.New(`{"error":"incorrect email input","details":"email already exists - conflict detected"}`)
	errInvalidPassword = errors.New(`{"error":"incorrect password input","details":"pass must have 8-256 chars and contain only ASCII"}`)
	errInvalidFullName = errors.New(`{"error":"incorrect fullName input","details":"fullName must have more than 3 chars"}`)
	errUsesNotExists   = errors.New(`{"error":"incorrect endpoint","details":"no user with such ID"}`)
)

type User struct {
	Id            string    `json:"id"`
	Email         string    `json:"email"`
	FullName      string    `json:"full-name"`
	Password      string    `json:"password"`
	CreatedAt     time.Time `json:"created-at"`
	LastUpdatedAt time.Time `json:"last-updated-at"`
}

func counter(i *int) string {
	*i++
	return fmt.Sprint(*i)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)     // get variables from request
	for _, v := range users { // iteration in DB
		if v.Id == params["id"] { //looking for user by ID in DB
			w.Header().Set("Content-Type", "application/json")   // if id exists, set header  and go out in line 55
			if err := json.NewEncoder(w).Encode(v); err != nil { // encode result into response and check for errors
				http.Error(w, err.Error(), http.StatusInternalServerError) //return error if something go wrong while encoding
				return                                                     //break function
			}
			return
		}
	}
	http.Error(w, errUsesNotExists.Error(), http.StatusNotFound)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if err := user.newUserEmailValidator(); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := user.newUserNameValidator(); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := user.newUserPassValidator(); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := compareEmail(users, user); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	user.Id = counter(&idCount)
	user.CreatedAt = time.Now()
	user.LastUpdatedAt = time.Now()
	users = append(users, user)
	json.NewEncoder(w).Encode(user)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if err := user.newUserEmailValidator(); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := user.newUserNameValidator(); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := user.newUserPassValidator(); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := compareEmail(users, user); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	params := mux.Vars(r)
	for i, v := range users {
		if v.Id == params["id"] {
			users = append(users[:i], users[i+1:]...)
			user.Id = params["id"]
			user.CreatedAt = v.CreatedAt
			user.LastUpdatedAt = time.Now()
			users = append(users, user)
			json.NewEncoder(w).Encode(user)
			return
		}
	}
	http.Error(w, errUsesNotExists.Error(), http.StatusNotFound)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for i, v := range users {
		if v.Id == params["id"] {
			users = append(users[:i], users[i+1:]...)
			return
		}
	}
	http.Error(w, errUsesNotExists.Error(), http.StatusNotFound)
}

func (u User) newUserNameValidator() error { // nil if ok
	if len(u.FullName) < 3 {
		return errInvalidFullName
	}
	return nil
}

func (u User) newUserEmailValidator() error { // nil if ok
	if len(u.Email) < 5 || len(u.Email) > 256 {
		return errInvalidEmail
	}
	for _, v := range u.Email {
		if v == '@' {
			return nil
		}
	}
	return errInvalidEmail
}

func (u User) newUserPassValidator() error { // nil if ok
	if len(u.Password) <= 7 || len(u.Password) > 256 {
		return errInvalidPassword
	} else {
		for _, v := range u.Password {
			if v < '!' || v > '~' {
				return errInvalidPassword
			}
		}
	}
	return nil
}

func compareEmail(us []User, u User) error { // nil if ok
	for _, v := range us {
		if u.Email == v.Email {
			return errConflictEmail
		}
	}
	return nil
}

func main() {
	users = append(users, User{
		Id:            counter(&idCount),
		Email:         "firstEmail@gmail.com",
		FullName:      "First Usr",
		Password:      "qwerty",
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
		//Deleted:       false,
	})

	users = append(users, User{
		Id:            counter(&idCount),
		Email:         "secondEmail@gmail.com",
		FullName:      "Second Usr",
		Password:      "qwerty",
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
		//Deleted:       false,
	})

	r := mux.NewRouter()
	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/users/", getUsers).Methods("GET")
	r.HandleFunc("/users/{id}", getUser).Methods("GET")
	r.HandleFunc("/users", createUser).Methods("POST")
	r.HandleFunc("/users/", createUser).Methods("POST")
	r.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	r.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", r))
}
