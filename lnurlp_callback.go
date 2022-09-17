package main

import (
	"os"
        log "github.com/sirupsen/logrus"
        "github.com/gorilla/mux"
        "net/http"
	"strconv"
)

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

        domain := os.Getenv("HOST_DOMAIN")
        if r.Host != domain {
                log.Warn("wrong host domain")
                write_error(w)
                return
        }

//TODO add err
        amount_msat, _ := strconv.ParseInt(amount, 10, 64)
        amount_sat      := amount_msat / 1000;

//TODO add err
        metadata := "[[\"text/identifier\",\"" + name + "@" + domain + "\"],[\"text/plain\",\"bolt card deposit\"]]"
        pr, r_hash, _ := add_invoice(amount_sat, metadata)

        jsonData := []byte(`{` +
                `"status":"OK",` +
                `"routes":[],` +
		`"pr":"` + pr + `"` +
        `}`)

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write(jsonData)

	go monitor_invoice_state(r_hash)
}
