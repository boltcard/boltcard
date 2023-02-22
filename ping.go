package main

import (
	"net/http"
)

func external_ping(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(`{"status":"OK","pong":"external API"}`)
	w.Write(jsonData)
}
