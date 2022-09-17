package main

import (
	"os"
        log "github.com/sirupsen/logrus"
        "github.com/gorilla/mux"
        "net/http"
	"strconv"
	"encoding/hex"
)

func lnurlp_callback(w http.ResponseWriter, r *http.Request) {
        name := mux.Vars(r)["name"]
        amount := r.URL.Query().Get("amount")

	card_id, err := db_get_card_id_for_name(name)
	if err != nil {
		log.Info("card name not found")
		write_error(w)
		return
	}

        log.WithFields(
                log.Fields{
                        "url_path": r.URL.Path,
                        "name": name,
			"card_id": card_id,
                        "amount": amount,
                        "req.Host": r.Host,
                },).Info("lnurlp_callback")

        domain := os.Getenv("HOST_DOMAIN")
        if r.Host != domain {
                log.Warn("wrong host domain")
                write_error(w)
                return
        }

        amount_msat, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
                log.Warn("amount is not a valid integer")
                write_error(w)
                return
        }

        amount_sat      := amount_msat / 1000;

        metadata := "[[\"text/identifier\",\"" + name + "@" + domain + "\"],[\"text/plain\",\"bolt card deposit\"]]"
        pr, r_hash, err := add_invoice(amount_sat, metadata)
	if err != nil {
                log.Warn("could not add_invoice")
                write_error(w)
                return
        }

	err = db_insert_receipt(card_id, pr, hex.EncodeToString(r_hash), amount_msat)
	if err != nil {
		log.Warn(err)
		write_error(w)
		return
	}

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
