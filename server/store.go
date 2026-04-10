package main

import (
	"github.com/jmoiron/sqlx"
)

func initDb(driver, connectionString string) *sqlx.DB {
	db, err := sqlx.Connect(driver, connectionString)

	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return db
}
