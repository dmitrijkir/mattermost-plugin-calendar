package main

import "github.com/jmoiron/sqlx"

var db *sqlx.DB


func initDb(driver, connectionString string)  {
    var err error
    db, err = sqlx.Connect(driver, connectionString)

    if err != nil {

    }

    db.MustExec(sqlSchema)
}

func GetDb() *sqlx.DB {
    return db
}