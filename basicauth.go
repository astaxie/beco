package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type htpasswd map[string]string

func NewBasicAuth(pw htpasswd) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		return BasicAuth{h, pw}
	}
	return fn
}

type BasicAuth struct {
	h  http.Handler
	pw htpasswd
}

func (b BasicAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	unauthed := func(url *url.URL) {
		log.Printf("Unauthorized access to %s", url)
		w.Header().Add("WWW-Authenticate", "basic realm=\"beco auth\"")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Not Authorized.")
	}
	auth, ok := r.Header["Authorization"]
	if !ok {
		unauthed(r.URL)
		return
	}
	encoded := strings.Split(auth[0], " ")
	if len(encoded) != 2 || encoded[0] != "Basic" {
		log.Printf("Error Authorization %q", auth)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded[1])
	if err != nil {
		log.Printf("Cannot decode %q: %s", auth, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	credentials := string(decoded)
	for method, user := range b.pw {
		if method == "r" && r.Method != "GET" {
			continue
		}
		if credentials == user {
			b.h.ServeHTTP(w, r)
			return
		}
	}
	unauthed(r.URL)
}
