package main

import "net/http"

var authProvider map[string]IAuth

type IAuth interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

func RegisterAuth(name string, auth IAuth) error {
	return nil
}
