package main

import (
	"EPI_innovation_2018/server/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/user", a.createUser).Methods("POST")
	a.Router.HandleFunc("/user", a.login).Methods("GET")
	a.Router.HandleFunc("/users", a.getUsers).Methods("GET")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.getUser).Methods("GET")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.updateUser).Methods("PUT")
	a.Router.HandleFunc("/user/{id:[0-9]+}", a.deleteUser).Methods("DELETE")
}

func (a *App) Initialize(user, password, dbname string) {
	var err error
	connectionString := "root:root@tcp(localhost:3306)/mysql"

	if a.DB, err = sql.Open("mysql", connectionString); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
		log.Fatal(err)
	}
	file, err := os.Open("server/db/exemple.sql")
	if err != nil {
		fmt.Printf("Open file error: %s\n", err.Error())
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	requests := strings.Split(string(content), ";\n")
	for _, request := range requests {
		if _, err := a.DB.Exec(request); err != nil {
			fmt.Printf("Connection error: %s\n", err.Error())
		}
	}
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var u user
	var cpy user
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&u); err != nil {
		fmt.Printf("Invalid request payload : %s, %d", w, http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	cpy = u
	if err := u.login(a.DB); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "User not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	if utils.CheckPasswordHash(cpy.Password, u.Password) == false {
		respondWithError(w, http.StatusNotFound, "Wrong password or email")
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	u := user{ID: id}
	if err := u.getUser(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "User not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var u user
	var err error
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		fmt.Printf("Invalid request payload : %s, %d", w, http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	fmt.Printf("User: %s // %s\n", u.Email, u.Password)
	if u.Password, err = utils.HashPassword(u.Password); err != nil {
		fmt.Printf("Error Hash password: %s\n", err.Error())
	}
	fmt.Printf("User: %s // %s\n", u.Email, u.Password)
	defer r.Body.Close()
	if err := u.createUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, u)
	fmt.Printf("Success %s\n", r.Form)
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	var u user
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close()
	u.ID = id
	if err := u.updateUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid User ID")
		return
	}
	u := user{ID: id}
	if err := u.deleteUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))
	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}
	users, err := getUsers(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, users)
}

func (a *App) Run(addr string) {
	fmt.Printf("Running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))
}
