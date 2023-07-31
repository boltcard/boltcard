package main

import (
	"github.com/boltcard/boltcard/db"
	"github.com/boltcard/boltcard/internalapi"
	"github.com/boltcard/boltcard/lnurlp"
	"github.com/boltcard/boltcard/lnurlw"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
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

	var external_router = mux.NewRouter()
	var internal_router = mux.NewRouter()

	// external API

	// ping
	external_router.Path("/ping").Methods("GET").HandlerFunc(external_ping)
	// createboltcard
	external_router.Path("/new").Methods("GET").HandlerFunc(new_card_request)
	// lnurlw for pos
	external_router.Path("/ln").Methods("GET").HandlerFunc(lnurlw.Response)
	external_router.Path("/cb").Methods("GET").HandlerFunc(lnurlw.Callback)
	// lnurlp for lightning address
	external_router.Path("/.well-known/lnurlp/{name}").Methods("GET").HandlerFunc(lnurlp.Response)
	external_router.Path("/lnurlp/{name}").Methods("GET").HandlerFunc(lnurlp.Callback)

	// internal API
	// this has no authentication and is not to be exposed publicly
	// it exists for use on a private virtual network within a docker container

	internal_router.Path("/ping").Methods("GET").HandlerFunc(internalapi.Internal_ping)
	internal_router.Path("/createboltcard").Methods("GET").HandlerFunc(internalapi.Createboltcard)
	internal_router.Path("/updateboltcard").Methods("GET").HandlerFunc(internalapi.Updateboltcard)
	internal_router.Path("/updateboltcardwithpin").Methods("GET").HandlerFunc(internalapi.Updateboltcardwithpin)
	internal_router.Path("/wipeboltcard").Methods("GET").HandlerFunc(internalapi.Wipeboltcard)
	internal_router.Path("/getboltcard").Methods("GET").HandlerFunc(internalapi.Getboltcard)

	port := db.Get_setting("HOST_PORT")
	if port == "" {
		port = "9000"
	}

	external_server := &http.Server{
		Handler:      external_router,
		Addr:         ":" + port, // consider adding host
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	internal_server := &http.Server{
		Handler:      internal_router,
		Addr:         ":9001",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}

	go external_server.ListenAndServe()
	go internal_server.ListenAndServe()

	select {}
}
