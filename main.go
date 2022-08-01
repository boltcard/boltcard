package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {
	log_level := os.Getenv("LOG_LEVEL")

	if log_level == "DEBUG" {
		log.SetLevel(log.DebugLevel)
	}

	log.SetFormatter(&log.JSONFormatter{
		DisableHTMLEscape: true,
	})

	mux := http.NewServeMux()

	mux.HandleFunc("/ln", lnurlw_response)
	mux.HandleFunc("/cb", lnurlw_callback)

	err := http.ListenAndServe(":9000", mux)
	log.Fatal(err)
}
