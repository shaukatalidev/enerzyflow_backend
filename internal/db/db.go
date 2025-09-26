package db

import (
    "database/sql"
    "log"
    _ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) {
    var err error
    DB, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatalf("failed to open DB: %v", err)
    }

    if err = DB.Ping(); err != nil {
        log.Fatalf("failed to ping DB: %v", err)
    }

    log.Println("SQLite DB connected successfully")
}