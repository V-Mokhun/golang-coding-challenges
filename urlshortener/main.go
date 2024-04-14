package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"strings"
)

func main() {
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
