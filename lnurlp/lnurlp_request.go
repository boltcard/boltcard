package lnurlp

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"github.com/boltcard/boltcard/db"
	"github.com/boltcard/boltcard/resp_err"
)

func Response(w http.ResponseWriter, r *http.Request) {
	if db.Get_setting("FUNCTION_LNURLP") != "ENABLE" {
		log.Debug("LNURLp function is not enabled")
		return
	}

	name := mux.Vars(r)["name"]

	log.WithFields(
		log.Fields{
			"url_path": r.URL.Path,
			"name":     name,
			"r.Host":   r.Host,
		}).Info("lnurlp_response")

	// look up domain setting (HOST_DOMAIN)

	domain := db.Get_setting("HOST_DOMAIN")
	if r.Host != domain {
		log.Warn("wrong host domain")
		resp_err.Write(w)
		return
	}

	// look up name in database (table cards, field card_name)

	card_count, err := db.Get_card_count_for_name_lnurlp(name)
	if err != nil {
		log.Warn("could not get card count for name")
		resp_err.Write(w)
		return
	}

	if card_count != 1 {
		log.Info("not one enabled card with that name")
		resp_err.Write(w)
		return
	}

	metadata := "[[\\\"text/identifier\\\",\\\"" + name + "@" + domain + "\\\"],[\\\"text/plain\\\",\\\"bolt card deposit\\\"]]"

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
