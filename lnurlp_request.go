package main

import (
	"os"
        log "github.com/sirupsen/logrus"
        "github.com/gorilla/mux"
        "net/http"
	"strconv"
)

func lnurlp_response(w http.ResponseWriter, r *http.Request) {
        name := mux.Vars(r)["name"]

        log.WithFields(
                log.Fields{
                        "url_path": r.URL.Path,
                        "name": name,
                        "r.Host": r.Host,
                },).Info("lnurlp_response")

// look up domain in env vars (HOST_DOMAIN)

        domain := os.Getenv("HOST_DOMAIN")
        if r.Host != domain {
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

        metadata := "[[\\\"text/identifier\\\",\\\"" + name + "@" + domain + "\\\"],[\\\"text/plain\\\",\\\"" + name + "@" + domain + "\\\"]]"

        jsonData := []byte(`{"status":"OK",` +
                `"callback":"https://` + domain + `/lnurlp/` + name + `",` +
                `"tag":"payRequest",` +
                `"maxSendable":1000000000,` +
                `"minSendable":1000,` +
                `"metadata":"` + metadata + `",` +
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
        metadata := "[[\"text/identifier\",\"" + name + "@" + domain + "\"],[\"text/plain\",\"" + name + "@" + domain + "\"]]"
        pr, _ := add_invoice(amount_sat, metadata)

        jsonData := []byte(`{` +
                `"status":"OK","successAction":{"tag":"message","message":"Payment success!"}` +
                `,"routes":[],"pr":"` + pr + `","disposable":false` +
        `}`)

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write(jsonData)
}
