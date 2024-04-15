package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
)

type DBUrl struct {
	key       string
	short_url string
	long_url  string
}

const KEY_LENGTH int = 8

var db *sql.DB

func main() {
	connectToDB()
	defer db.Close()

	res, err := shortenUrl("https://www.google.com/search?q=hash&oq=something#something")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
}

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
	if err := row.Scan(&url.key, &url.long_url, &url.short_url); err != nil {
		if err == sql.ErrNoRows {
			return url, sql.ErrNoRows
		}
		return url, fmt.Errorf("unexpected error")
	}

	return url, nil
}

func shortenUrl(urlToShorten string) (string, error) {
	key := hashUrl(urlToShorten)
	shortenedKey := key[:KEY_LENGTH]
	_, err := queryUrl(shortenedKey)

	// if key does not exist in db
	if err != nil {
		if err == sql.ErrNoRows {
			_, insertErr := db.Exec("INSERT INTO url (`key`, long_url, short_url) VALUES(?, ?, ?)", shortenedKey, urlToShorten, "http://localhost/"+shortenedKey)

			// TODO: handle insertion error http
			if insertErr != nil {
				return shortenedKey, insertErr
			}

			fmt.Println("inserted key: " + shortenedKey)

			// TODO: handle http 200 status
			return shortenedKey, nil
		} else {
			// TODO: handle http status when error
			return shortenedKey, err
		}
	}

	// TODO: handle url already exists
	return shortenedKey, nil
}

func hashUrl(urlToHash string) string {
	hash := sha256.Sum256([]byte(constructUrl(urlToHash)))
	return hex.EncodeToString(hash[:])
}

func constructUrl(urlToConstruct string) string {
	u, err := url.Parse(urlToConstruct)

	if err != nil {
		log.Fatal(err)
	}

	hostname := u.Hostname()
	var newUrl string
	if strings.HasPrefix(hostname, "www.") {
		newUrl = hostname[4:]
	} else {
		newUrl = hostname
	}
	newUrl += u.EscapedPath()
	if u.RawQuery != "" {
		newUrl += "?" + u.RawQuery
	}
	if u.EscapedFragment() != "" {
		newUrl += "#" + u.EscapedFragment()
	}

	return newUrl
}
