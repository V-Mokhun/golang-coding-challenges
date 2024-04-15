package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			msg := "Content-Type header is not application/json"
			return &malformedRequest{status: http.StatusUnsupportedMediaType, msg: msg}
		}
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprint("Request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
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
			return insertUrl(newUrl)
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
