package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
	"net/http"
	"time"
	"os"
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

func lnurlp_response(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	log.WithFields(
        	log.Fields{
            		"url_path": r.URL.Path,
			"name": name,
			"r.Host": r.Host,
		},).Info("lnurlp_response")

// look up domain in env vars (HOST_DOMAIN)
	env_host_domain := os.Getenv("HOST_DOMAIN")
	if r.Host != env_host_domain {
		log.Warn("wrong host domain")
		write_error(w)
		return
	}

// look up name in database (table cards, field card_name)

        card_count, err := db_get_card_count_for_name(name)
        if err != nil {
	        log.Warn("could not get card count for name")
		write_error(w)
		return
        }

	if card_count != 1 {
		log.Info("not one card with that name")
		write_error(w)
		return
	}

	jsonData := []byte(`{"status":"OK",` +
		`"callback":"https://` + env_host_domain + `/lnurlp/` + name + `",` +
		`"tag":"payRequest",` +
		`"maxSendable":1000000000,` +
		`"minSendable":1000,` +
		`"metadata":"[[\"text/plain\",\"` + name + `@` + env_host_domain + `\"]]",` +
		`"commentAllowed":0` +
	`}`)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func lnurlp_callback(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	amount := r.URL.Query().Get("amount");

	log.WithFields(
        	log.Fields{
            		"url_path": r.URL.Path,
			"name": name,
			"amount": amount,
			"req.Host": r.Host,
		},).Info("lnurlp_callback")


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

// createboltcard
	router.Path("/new").Methods("GET").HandlerFunc(new_card_request)
// lnurlw for pos
	router.Path("/ln").Methods("GET").HandlerFunc(lnurlw_response)
	router.Path("/cb").Methods("GET").HandlerFunc(lnurlw_callback)
// lnurlp for lightning address lnurlp
	router.Path("/.well-known/lnurlp/{name}").Methods("GET").HandlerFunc(lnurlp_response)
	router.Path("/lnurlp/{name}").Methods("GET").HandlerFunc(lnurlp_callback)

	port := os.Getenv("HOST_PORT")
	if len(port) == 0 {
		port = "9000"
	}

	srv := &http.Server {
		Handler:      router,
		Addr:         ":" + port, // consider adding host
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	srv.ListenAndServe()
}
