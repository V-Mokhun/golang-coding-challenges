package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
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

	http.HandleFunc("/", handleRedirectAndDelete)
	http.HandleFunc("/shorten", handleShorten)
	http.HandleFunc("/view", handleView)

	http.ListenAndServe(":8080", nil)
}

func handleShorten(w http.ResponseWriter, r *http.Request) {
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
}

func handleRedirectAndDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodDelete {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Path[1:]
	dbUrl, err := queryUrl(key)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodDelete {
		_, err := deleteUrl(key)
		if err != nil {
			http.Error(w, "Could not delete url", http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Location", dbUrl.LongUrl)
	http.Redirect(w, r, dbUrl.LongUrl, http.StatusFound)
}

func handleView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	fileName := "templates/index.html"

	t, _ := template.ParseFiles(fileName)
	err := t.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
