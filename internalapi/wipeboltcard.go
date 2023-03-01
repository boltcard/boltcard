package internalapi

import (
	"github.com/boltcard/boltcard/db"
	"github.com/boltcard/boltcard/resp_err"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func Wipeboltcard(w http.ResponseWriter, r *http.Request) {
	if db.Get_setting("FUNCTION_INTERNAL_API") != "ENABLE" {
		msg := "wipeboltcard: internal API function is not enabled"
		log.Debug(msg)
		resp_err.Write_message(w, msg)
		return
	}

	card_name := r.URL.Query().Get("card_name")

	// check if card_name has been given

	if card_name == "" {
		msg := "wipeboltcard: the card name must be set"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	// set the card as wiped and disabled, get the keys

	card_wipe_info_values, err := db.Wipe_card(card_name)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	// log the request

	log.WithFields(log.Fields{
		"card_name": card_name}).Info("wipeboltcard API request")

	// generate a response

	jsonData := `{"status":"OK",` +
		`"action": "wipe",` +
		`"id": ` + strconv.Itoa(card_wipe_info_values.Id) + `,` +
		`"k0": "` + card_wipe_info_values.K0 + `",` +
		`"k1": "` + card_wipe_info_values.K1 + `",` +
		`"k2": "` + card_wipe_info_values.K2 + `",` +
		`"k3": "` + card_wipe_info_values.K3 + `",` +
		`"k4": "` + card_wipe_info_values.K4 + `",` +
		`"uid": "` + card_wipe_info_values.Uid + `",` +
		`"version": 1}`

	// log the response

	log.WithFields(log.Fields{
		"card_name": card_name, "response": jsonData}).Info("wipeboltcard API response")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonData))
}
