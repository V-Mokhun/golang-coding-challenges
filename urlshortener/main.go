package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
)

type DBUrl struct {
	Key      string `json:"key"`
	ShortUrl string `json:"shortUrl"`
	LongUrl  string `json:"longUrl"`
}

type ApiURL struct {
	Url string `json:"url"`
}

const KEY_LENGTH int = 8

var db *sql.DB

func main() {
	connectToDB()

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		var url ApiURL
		err := decodeJSONBody(w, r, &url)
		if err != nil {
			var mr *malformedRequest
			if errors.As(err, &mr) {
				http.Error(w, mr.msg, mr.status)
			} else {
				log.Print(err.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		shortenedUrl, err := shortenUrl(url.Url)
		if err != nil {
			log.Print("Shorten url error:", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		shortenedUrlJson, err := json.Marshal(shortenedUrl)
		if err != nil {
			log.Print("Marshal error:", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(shortenedUrlJson)
	})

	http.ListenAndServe(":8080", nil)
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
	if err := row.Scan(&url.Key, &url.LongUrl, &url.ShortUrl); err != nil {
		if err == sql.ErrNoRows {
			return url, sql.ErrNoRows
		}
		return url, fmt.Errorf("unexpected error")
	}

	return url, nil
}

func shortenUrl(urlToShorten string) (DBUrl, error) {
	key := hashUrl(urlToShorten)
	shortenedKey := key[:KEY_LENGTH]
	existingUrl, err := queryUrl(shortenedKey)

	// if key does not exist in db
	if err != nil {
		newUrl := DBUrl{
			Key:      shortenedKey,
			ShortUrl: "http://localhost/" + shortenedKey,
			LongUrl:  urlToShorten,
		}

		if err == sql.ErrNoRows {
			_, insertErr := db.Exec("INSERT INTO url (`key`, long_url, short_url) VALUES(?, ?, ?)", newUrl.Key, newUrl.LongUrl, newUrl.ShortUrl)

			if insertErr != nil {
				return newUrl, insertErr
			}

			return newUrl, nil
		} else {
			return newUrl, err
		}
	}

	return existingUrl, nil
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
