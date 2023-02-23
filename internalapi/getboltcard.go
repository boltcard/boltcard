package internalapi

import (
	"net/http"
	"strconv"

	"github.com/boltcard/boltcard/db"
	"github.com/boltcard/boltcard/resp_err"
	log "github.com/sirupsen/logrus"
)

func Getboltcard(w http.ResponseWriter, r *http.Request) {
	if db.Get_setting("FUNCTION_INTERNAL_API") != "ENABLE" {
		msg := "getboltcard: internal API function is not enabled"
		log.Debug(msg)
		resp_err.Write_message(w, msg)
		return
	}
	card_name := r.URL.Query().Get("card_name")

	// check if card_name exists

	card_count, err := db.Get_card_name_count(card_name)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	if card_count == 0 {
		msg := "getboltcard: the card name does not exist in the database"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	// log the request

	log.WithFields(log.Fields{
		"card_name": card_name}).Info("getboltcard API request")

	// get the card record

	c, err := db.Get_card_from_card_name(card_name)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	jsonData := []byte(`{"status":"OK",` +
		`"uid": "` + c.Db_uid + `",` +
		`"last_counter_value": "` + strconv.FormatUint(uint64(c.Last_counter_value), 10) + `",` +
		`"lnurlw_enable": "` + c.Lnurlw_enable + `",` +
		`"tx_limit_sats": "` + strconv.Itoa(c.Tx_limit_sats) + `",` +
		`"day_limit_sats": "` + strconv.Itoa(c.Day_limit_sats) + `"}`)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
