package main

import (
	"database/sql"
	"fmt"
)

type user struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *user) getUser(db *sql.DB) error {
	statement := fmt.Sprintf("SELECT email, password FROM users WHERE id=%d", u.ID)
	return db.QueryRow(statement).Scan(&u.Email, &u.Password)
}
func (u *user) login(db *sql.DB) error {
	statement := fmt.Sprintf("SELECT email, password, id FROM users WHERE email='%s'", u.Email)
	return db.QueryRow(statement).Scan(&u.Email, &u.Password, &u.ID)
}
func (u *user) updateUser(db *sql.DB) error {
	statement := fmt.Sprintf("UPDATE users SET email='%s', password='%s' WHERE id=%d", u.Email, u.Password, u.ID)
	_, err := db.Exec(statement)
	return err
}
func (u *user) deleteUser(db *sql.DB) error {
	statement := fmt.Sprintf("DELETE FROM users WHERE id=%d", u.ID)
	_, err := db.Exec(statement)
	return err
}
func (u *user) createUser(db *sql.DB) error {
	//TODO vérifier e-mail
	fmt.Printf("Create User next\n")
	statement := fmt.Sprintf("INSERT INTO users(email, password) VALUES('%s', '%s')", u.Email, u.Password)
	_, err := db.Exec(statement)
	if err != nil {
		return err
	}
	err = db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&u.ID)
	if err != nil {
		return err
	}
	return nil
}
func getUsers(db *sql.DB, start, count int) ([]user, error) {
	//TODO when delete decrease id
	statement := fmt.Sprintf("SELECT id, email, password FROM users LIMIT %d OFFSET %d", count, start)
	rows, err := db.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := []user{}
	for rows.Next() {
		var u user
		if err := rows.Scan(&u.ID, &u.Email, &u.Password); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
