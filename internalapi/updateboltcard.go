package internalapi

import (
	"github.com/boltcard/boltcard/db"
	"github.com/boltcard/boltcard/resp_err"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func Updateboltcard(w http.ResponseWriter, r *http.Request) {
	if db.Get_setting("FUNCTION_INTERNAL_API") != "ENABLE" {
		msg := "updateboltcard: internal API function is not enabled"
		log.Debug(msg)
		resp_err.Write_message(w, msg)
		return
	}

	tx_max_str := r.URL.Query().Get("tx_max")
	tx_max, err := strconv.Atoi(tx_max_str)
	if err != nil {
		msg := "updateboltcard: tx_max is not a valid integer"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	enable_flag_str := r.URL.Query().Get("enable")
	enable_flag, err := strconv.ParseBool(enable_flag_str)
	if err != nil {
		msg := "updateboltcard: enable is not a valid boolean"
		log.Warn(msg)
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
		msg := "updateboltcard: the card name does not exist in the database"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	// log the request

	log.WithFields(log.Fields{
		"card_name": card_name, "tx_max": tx_max,
		"enable": enable_flag}).Info("createboltcard API request")

	// update the card record

	err = db.Update_card(card_name, tx_max, enable_flag)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	// send a response

	jsonData := []byte(`{"status":"OK"}`)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
