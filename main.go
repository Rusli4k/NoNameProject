package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const connStr string = "user=postgres password=123 dbname=nnm sslmode=disable"

type customError struct {
	Error     interface{}
	Details   interface{}
	TimeStamp time.Time
}

var (
	errInvalidEmail    = customError{Error: "incorrect email input", Details: "email must have 5-256 chars and contain @"}
	errConflictEmail   = customError{Error: "incorrect email input", Details: "email already exists - conflict detected"}
	errInvalidPassword = customError{Error: "incorrect password input", Details: "pass must have 8-256 chars and contain only ASCII"}
	errInvalidFullName = customError{Error: "incorrect fullName input", Details: "fullName must have more than 3 chars"}
	errUsesNotExists   = customError{Error: "incorrect endpoint", Details: "no user with such ID"}
)

type User struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	FullName      string `json:"full-name"`
	Password      string `json:"password"`
	CreatedAt     string `json:"created-at"`
	LastUpdatedAt string `json:"last-updated-at"`
}

func convertErrToCustomError(e error) (t customError) {
	t.Error = e
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

func getUsers(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("select * FROM userz ORDER BY id")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		u := User{}
		err := rows.Scan(&u.Id, &u.Email, &u.FullName, &u.Password, &u.CreatedAt, &u.LastUpdatedAt)
		if err != nil {
			log.Fatal(err)
			continue
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	var user User

	row := db.QueryRow("SELECT * FROM userz WHERE id=$1", params["id"])
	err = row.Scan(&user.Id, &user.Email, &user.FullName, &user.Password, &user.CreatedAt, &user.LastUpdatedAt)
	if err != nil {
		sendCustomErrorToHttp(w, http.StatusNotFound, errUsesNotExists)
		sendCustomErrorToHttp(w, http.StatusNotFound, convertErrToCustomError(err))
		return
	}
	if err := json.NewEncoder(w).Encode(user); err != nil {
		sendCustomErrorToHttp(w, http.StatusInternalServerError, convertErrToCustomError(err))
		return
	}
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM userz")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var user User
	var users []User

	for rows.Next() {
		u := User{}
		err := rows.Scan(&u.Id, &u.Email, &u.FullName, &u.Password, &u.CreatedAt, &u.LastUpdatedAt)
		if err != nil {
			log.Fatal(err)
			continue
		}
		users = append(users, u)
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		sendCustomErrorToHttp(w, http.StatusUnsupportedMediaType, convertErrToCustomError(err))
		return
	}

	if err := user.newUserEmailValidator(); err != nil {
		sendCustomErrorToHttp(w, http.StatusUnsupportedMediaType, errInvalidEmail)
		return
	}

	if err := user.newUserNameValidator(); err != nil {
		sendCustomErrorToHttp(w, http.StatusUnsupportedMediaType, errInvalidFullName)
		return
	}

	if err := user.newUserPassValidator(); err != nil {
		sendCustomErrorToHttp(w, http.StatusUnprocessableEntity, errInvalidPassword)
		return
	}

	if err := compareEmail(users, user); err != nil {
		sendCustomErrorToHttp(w, http.StatusConflict, errConflictEmail)
		return
	}

	var i int
	row := db.QueryRow("SELECT MAX(id) FROM userz")
	row.Scan(&i)

	user.Id = fmt.Sprint(i + 1)
	user.CreatedAt = time.Now().String()
	user.LastUpdatedAt = time.Now().String()

	_, err = db.Exec("INSERT INTO userz (id, email, fullname, password, createdat, lastupdatedat) VALUES ($1,$2,$3,$4,$5,$6)", user.Id, user.Email, user.FullName, user.Password, user.CreatedAt, user.LastUpdatedAt)
	if err != nil {
		sendCustomErrorToHttp(w, http.StatusInternalServerError, convertErrToCustomError(err))
	}

	json.NewEncoder(w).Encode(user)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM userz")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var user User
	var users []User

	for rows.Next() {
		u := User{}
		err := rows.Scan(&u.Id, &u.Email, &u.FullName, &u.Password, &u.CreatedAt, &u.LastUpdatedAt)
		if err != nil {
			log.Fatal(err)
			continue
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	for _, v := range users {
		if v.Id == params["id"] {
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				sendCustomErrorToHttp(w, http.StatusUnsupportedMediaType, convertErrToCustomError(err))
				return
			}

			if err := user.newUserEmailValidator(); err != nil {
				sendCustomErrorToHttp(w, http.StatusUnprocessableEntity, errInvalidEmail)
				return
			}

			if err := user.newUserNameValidator(); err != nil {
				sendCustomErrorToHttp(w, http.StatusUnprocessableEntity, errInvalidFullName)
				return
			}

			if err := user.newUserPassValidator(); err != nil {
				sendCustomErrorToHttp(w, http.StatusUnprocessableEntity, errInvalidPassword)
				return
			}

			if err := compareEmail(users, user); err != nil {
				sendCustomErrorToHttp(w, http.StatusConflict, errConflictEmail)
				return
			}

			user.Id = params["id"]
			user.CreatedAt = v.CreatedAt
			user.LastUpdatedAt = time.Now().String()

			_, err = db.Exec("UPDATE userz SET email = $2, fullname =$3, password =$4,	createdat =$5,lastupdatedat =$6 WHERE id= $1", user.Id, user.Email, user.FullName, user.Password, user.CreatedAt, user.LastUpdatedAt)

			if err != nil {
				sendCustomErrorToHttp(w, http.StatusInternalServerError, convertErrToCustomError(err))
			}

			json.NewEncoder(w).Encode(user)
			return
		}
	}
	sendCustomErrorToHttp(w, http.StatusNotFound, errUsesNotExists)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM userz")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		u := User{}
		err := rows.Scan(&u.Id, &u.Email, &u.FullName, &u.Password, &u.CreatedAt, &u.LastUpdatedAt)
		if err != nil {
			log.Fatal(err)
			continue
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, v := range users {
		if v.Id == params["id"] {
			_, err = db.Exec("delete FROM userz WHERE id = $1", v.Id)
			if err != nil {
				sendCustomErrorToHttp(w, http.StatusInternalServerError, convertErrToCustomError(err))
			}
			return
		}
	}
	sendCustomErrorToHttp(w, http.StatusNotFound, errUsesNotExists)
}

func (u User) newUserNameValidator() error {
	if len(u.FullName) < 3 {
		return errors.New("errInvalidFullName")
	}
	return nil
}

func (u User) newUserEmailValidator() error {
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

func (u User) newUserPassValidator() error {
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

func compareEmail(us []User, u User) error {
	for _, v := range us {
		if u.Email == v.Email {
			return errors.New("errConflictEmail")
		}
	}
	return nil
}

func main() {
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
