package db

import (
    "database/sql"
    "log"

    _ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func Connect(dbURL string) {
    var err error
    DB, err = sql.Open("pgx", dbURL)
    if err != nil {
        log.Fatalf("failed to connect to DB: %v", err)
    }

    err = DB.Ping()
    if err != nil {
        log.Fatalf("failed to ping DB: %v", err)
    }

    log.Println("Connected to Neon Postgres successfully!")
}

