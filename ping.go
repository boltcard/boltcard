package main

import (
	"net/http"
)

func external_ping(w http.ResponseWriter, req *http.Request) {
	ping(w, "external API")
}

func internal_ping(w http.ResponseWriter, req *http.Request) {
	ping(w, "internal API")
}

func ping(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(`{"status":"OK","pong":"` + message + `"}`)
	w.Write(jsonData)
}
