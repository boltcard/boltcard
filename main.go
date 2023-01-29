package main

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var router = mux.NewRouter()

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
	log_level := db_get_setting("LOG_LEVEL")

	switch log_level {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
		log.Info("bolt card service started - debug log level")
	case "PRODUCTION":
		log.Info("bolt card service started - production log level")
	default:
		// log.Fatal calls os.Exit(1) after logging the error
		log.Fatal("error getting a valid LOG_LEVEL setting from the database")
	}

	log.SetFormatter(&log.JSONFormatter{
		DisableHTMLEscape: true,
	})

	// createboltcard
	router.Path("/new").Methods("GET").HandlerFunc(new_card_request)
	// lnurlw for pos
	router.Path("/ln").Methods("GET").HandlerFunc(lnurlw_response)
	router.Path("/cb").Methods("GET").HandlerFunc(lnurlw_callback)
	// lnurlp for lightning address
	router.Path("/.well-known/lnurlp/{name}").Methods("GET").HandlerFunc(lnurlp_response)
	router.Path("/lnurlp/{name}").Methods("GET").HandlerFunc(lnurlp_callback)

	port := db_get_setting("HOST_PORT")
	if port == "" {
		port = "9000"
	}

	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + port, // consider adding host
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	srv.ListenAndServe()
}
