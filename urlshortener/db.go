package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

func connectToDB() {
	cfg := mysql.Config{
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "url_shortener",
	}
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
}

func queryUrl(key string) (DBUrl, error) {
	var url DBUrl

	row := db.QueryRow("SELECT * FROM url WHERE `key` = ?", key)
	if err := row.Scan(&url.Key, &url.LongUrl, &url.ShortUrl); err != nil {
		if err == sql.ErrNoRows {
			return url, sql.ErrNoRows
		}
		return url, fmt.Errorf("unexpected error")
	}

	return url, nil
}

func insertUrl(newUrl DBUrl) (DBUrl, error) {
	_, insertErr := db.Exec("INSERT INTO url (`key`, long_url, short_url) VALUES(?, ?, ?)", newUrl.Key, newUrl.LongUrl, newUrl.ShortUrl)

	if insertErr != nil {
		return newUrl, insertErr
	}

	return newUrl, nil
}

func deleteUrl(key string) (bool, error) {
	_, deleteErr := db.Exec("DELETE FROM url WHERE `key` = ?", key)

	if deleteErr != nil {
		return false, deleteErr
	}

	return true, nil
}
