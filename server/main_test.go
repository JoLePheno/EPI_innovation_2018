package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var a App

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS users(
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(50) NOT NULL,
    password VARCHAR(50) NOT NULL
)`

func TestMain(m *testing.M) {
	a = App{}
	a.Initialize("root", "root", "mysql")
	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM users")
	a.DB.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
}

func TestEmptyTable(t *testing.T) {
	clearTable()
	req, _ := http.NewRequest("GET", "/users", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetNonExistentUser(t *testing.T) {
	clearTable()
	req, _ := http.NewRequest("GET", "/user/45", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "User not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'User not found'. Got '%s'", m["error"])
	}
}

func addUsers(count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		statement := fmt.Sprintf("INSERT INTO users(email, password) VALUES('%s', '%s')", "fersse", "chien")
		a.DB.Exec(statement)
	}
}

func TestCreateUser(t *testing.T) {
	clearTable()
	payload := []byte(`{"email":"test user","password":blabla}`)
	req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	fmt.Printf("email: %s, password: %s\n", m["email"], m["password"])
	if m["email"] != "test user" {
		t.Errorf("Expected user name to be 'test user'. Got '%v'", m["email"])
	}
	if m["password"] != "blabla" {
		t.Errorf("Expected user age to be 'blabla'. Got '%s'", m["password"])
	}
	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected user ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetUser(t *testing.T) {
	clearTable()
	addUsers(1)
	req, _ := http.NewRequest("GET", "/user/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestDeleteUser(t *testing.T) {
	clearTable()
	addUsers(1)
	req, _ := http.NewRequest("GET", "/user/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	req, _ = http.NewRequest("DELETE", "/user/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	req, _ = http.NewRequest("GET", "/user/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}
