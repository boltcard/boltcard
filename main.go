package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func write_error(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(`{"status":"ERROR","reason":"bad request"}`)
	w.Write(jsonData)
}

func write_error_message(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(`{"status":"ERROR","reason":"` + message + `"}`)
	w.Write(jsonData)
}

func main() {
	log_level := os.Getenv("LOG_LEVEL")

	if log_level == "DEBUG" {
		log.SetLevel(log.DebugLevel)
		log.Info("bolt card service started - debug log level")
	} else {
		log.Info("bolt card service started - production log level")
	}

	log.SetFormatter(&log.JSONFormatter{
		DisableHTMLEscape: true,
	})

	mux := http.NewServeMux()

	mux.HandleFunc("/new", new_card_request)
	mux.HandleFunc("/ln", lnurlw_response)
	mux.HandleFunc("/cb", lnurlw_callback)

	err := http.ListenAndServe(":9000", mux)
	log.Fatal(err)
}
