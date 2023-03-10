package main

import "github.com/jmoiron/sqlx"

var db *sqlx.DB

func initDb(driver, connectionString string) *sqlx.DB {
	var err error
	db, err = sqlx.Connect(driver, connectionString)

	if err != nil {

	}

	return db
}
