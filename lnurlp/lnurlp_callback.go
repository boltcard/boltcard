package lnurlp

import (
	"encoding/hex"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"github.com/boltcard/boltcard/lnd"
	"github.com/boltcard/boltcard/db"
	"github.com/boltcard/boltcard/resp_err"
)

func Callback(w http.ResponseWriter, r *http.Request) {
	if db.Get_setting("FUNCTION_LNURLP") != "ENABLE" {
		log.Debug("LNURLp function is not enabled")
		return
	}

	name := mux.Vars(r)["name"]
	amount := r.URL.Query().Get("amount")

	card_id, err := db.Get_card_id_for_name(name)
	if err != nil {
		log.Info("card name not found")
		resp_err.Write(w)
		return
	}

	log.WithFields(
		log.Fields{
			"url_path": r.URL.Path,
			"name":     name,
			"card_id":  card_id,
			"amount":   amount,
			"req.Host": r.Host,
		}).Info("lnurlp_callback")

	domain := db.Get_setting("HOST_DOMAIN")
	if r.Host != domain {
		log.Warn("wrong host domain")
		resp_err.Write(w)
		return
	}

	amount_msat, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		log.Warn("amount is not a valid integer")
		resp_err.Write(w)
		return
	}

	amount_sat := amount_msat / 1000

	metadata := "[[\"text/identifier\",\"" + name + "@" + domain + "\"],[\"text/plain\",\"bolt card deposit\"]]"
	pr, r_hash, err := lnd.Add_invoice(amount_sat, metadata)
	if err != nil {
		log.Warn("could not add_invoice")
		resp_err.Write(w)
		return
	}

	err = db.Insert_receipt(card_id, pr, hex.EncodeToString(r_hash), amount_msat)
	if err != nil {
		log.Warn(err)
		resp_err.Write(w)
		return
	}

	go lnd.Monitor_invoice_state(r_hash)

	log.Debug("sending 'status OK' response")

	jsonData := []byte(`{` +
		`"status":"OK",` +
		`"routes":[],` +
		`"pr":"` + pr + `"` +
		`}`)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
