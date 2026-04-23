package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "series.db")
	if err != nil {
		panic(err)
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS series (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		name            TEXT    NOT NULL,
		current_episode INTEGER NOT NULL DEFAULT 0,
		total_episodes  INTEGER NOT NULL,
		image_url       TEXT    NOT NULL DEFAULT ''
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS ratings (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		series_id INTEGER UNIQUE NOT NULL,
		rating    INTEGER NOT NULL CHECK(rating >= 0 AND rating <= 10),
		FOREIGN KEY (series_id) REFERENCES series(id)
	)`)

	return db
}
