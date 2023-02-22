package internalapi

import (
	"net/http"
)

func Internal_ping(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(`{"status":"OK","pong":"internal API"}`)
	w.Write(jsonData)
}
