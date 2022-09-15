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

type customError struct {
	MainText  interface{}
	Details   interface{}
	TimeStamp time.Time
}

var (
	errInvalidEmail    = customError{MainText: "incorrect email input", Details: "email must have 5-256 chars and contain @"}
	errConflictEmail   = customError{MainText: "incorrect email input", Details: "email already exists - conflict detected"}
	errInvalidPassword = customError{MainText: "incorrect password input", Details: "pass must have 8-256 chars and contain only ASCII"}
	errInvalidFullName = customError{MainText: "incorrect fullName input", Details: "fullName must have more than 3 chars"}
	errUsesNotExists   = customError{MainText: "incorrect endpoint", Details: "no user with such ID"}
)

type User struct {
	Id            string    `json:"id"`
	Email         string    `json:"email"`
	FullName      string    `json:"full-name"`
	Password      string    `json:"password"`
	CreatedAt     time.Time `json:"created-at"`
	LastUpdatedAt time.Time `json:"last-updated-at"`
}

func convertErrToCustomError(e error) (t customError) {
	t.MainText = e
	t.TimeStamp = time.Now()
	return
}

func sendCustomErrorToHttp(w http.ResponseWriter, statusCode int, e customError) {
	e.TimeStamp = time.Now()
	jsErr, err := json.Marshal(e)
	if err != nil {
		log.Fatal(err)
	}
	w.WriteHeader(statusCode)
	w.Write(jsErr)
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
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, v := range users {
		if v.Id == params["id"] {
			if err := json.NewEncoder(w).Encode(v); err != nil {
				sendCustomErrorToHttp(w, http.StatusInternalServerError, convertErrToCustomError(err))
				//http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
	}
	sendCustomErrorToHttp(w, http.StatusNotFound, errUsesNotExists)
	// w.WriteHeader(http.StatusNotFound)
	// w.Write(errUsesNotExists)
	//http.Error(w, errUsesNotExists.Error(), http.StatusNotFound)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		sendCustomErrorToHttp(w, http.StatusUnsupportedMediaType, convertErrToCustomError(err))
		//http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if err := user.newUserEmailValidator(); err != nil {
		sendCustomErrorToHttp(w, http.StatusUnsupportedMediaType, errInvalidEmail)
		//http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := user.newUserNameValidator(); err != nil {
		sendCustomErrorToHttp(w, http.StatusUnsupportedMediaType, errInvalidFullName)
		//http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := user.newUserPassValidator(); err != nil {
		sendCustomErrorToHttp(w, http.StatusUnprocessableEntity, errInvalidPassword)
		//http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := compareEmail(users, user); err != nil {
		sendCustomErrorToHttp(w, http.StatusConflict, errConflictEmail)
		//http.Error(w, err.Error(), http.StatusConflict)
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
		sendCustomErrorToHttp(w, http.StatusUnsupportedMediaType, convertErrToCustomError(err))
		//http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	if err := user.newUserEmailValidator(); err != nil {
		sendCustomErrorToHttp(w, http.StatusUnprocessableEntity, errInvalidEmail)
		//		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := user.newUserNameValidator(); err != nil {
		sendCustomErrorToHttp(w, http.StatusUnprocessableEntity, errInvalidFullName)
		//		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := user.newUserPassValidator(); err != nil {
		sendCustomErrorToHttp(w, http.StatusUnprocessableEntity, errInvalidPassword)
		//		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := compareEmail(users, user); err != nil {
		sendCustomErrorToHttp(w, http.StatusConflict, errConflictEmail)
		//		http.Error(w, err.Error(), http.StatusConflict)
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
	sendCustomErrorToHttp(w, http.StatusNotFound, errUsesNotExists)
	//http.Error(w, errUsesNotExists.Error(), http.StatusNotFound)
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
	sendCustomErrorToHttp(w, http.StatusNotFound, errUsesNotExists)
	//http.Error(w, errUsesNotExists.Error(), http.StatusNotFound)
}

func (u User) newUserNameValidator() error { // nil if ok
	if len(u.FullName) < 3 {
		return errors.New("errInvalidFullName")
	}
	return nil
}

func (u User) newUserEmailValidator() error { // nil if ok
	if len(u.Email) < 5 || len(u.Email) > 256 {
		return errors.New("errInvalidEmail")
	}
	for _, v := range u.Email {
		if v == '@' {
			return nil
		}
	}
	return errors.New("errInvalidEmail")
}

func (u User) newUserPassValidator() error { // nil if ok
	if len(u.Password) <= 7 || len(u.Password) > 256 {
		return errors.New("errInvalidPassword")
	} else {
		for _, v := range u.Password {
			if v < '!' || v > '~' {
				return errors.New("errInvalidPassword")
			}
		}
	}
	return nil
}

func compareEmail(us []User, u User) error { // nil if ok
	for _, v := range us {
		if u.Email == v.Email {
			return errors.New("errConflictEmail")
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
	})

	users = append(users, User{
		Id:            counter(&idCount),
		Email:         "secondEmail@gmail.com",
		FullName:      "Second Usr",
		Password:      "qwerty",
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
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
