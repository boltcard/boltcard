package main

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
	"github.com/boltcard/boltcard/db"
	"github.com/boltcard/boltcard/lnurlw"
	"github.com/boltcard/boltcard/lnurlp"
)

var router = mux.NewRouter()

func main() {
	log_level := db.Get_setting("LOG_LEVEL")

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
	router.Path("/ln").Methods("GET").HandlerFunc(lnurlw.Response)
	router.Path("/cb").Methods("GET").HandlerFunc(lnurlw.Callback)
	// lnurlp for lightning address
	router.Path("/.well-known/lnurlp/{name}").Methods("GET").HandlerFunc(lnurlp.Response)
	router.Path("/lnurlp/{name}").Methods("GET").HandlerFunc(lnurlp.Callback)

	port := db.Get_setting("HOST_PORT")
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
