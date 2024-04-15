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

var db *sql.DB

func main() {
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
	defer db.Close()

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	fmt.Println(hashUrl("https://www.google.com/search?q=hash&oq=something#something"))
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
