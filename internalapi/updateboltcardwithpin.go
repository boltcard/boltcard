package internalapi

import (
	"github.com/boltcard/boltcard/db"
	"github.com/boltcard/boltcard/resp_err"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func Updateboltcardwithpin(w http.ResponseWriter, r *http.Request) {
	if db.Get_setting("FUNCTION_INTERNAL_API") != "ENABLE" {
		msg := "updateboltcardwithpin: internal API function is not enabled"
		log.Debug(msg)
		resp_err.Write_message(w, msg)
		return
	}

	enable_flag_str := r.URL.Query().Get("enable")
	enable_flag, err := strconv.ParseBool(enable_flag_str)
	if err != nil {
		msg := "updateboltcardwithpin: enable is not a valid boolean"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	tx_max_str := r.URL.Query().Get("tx_max")
	tx_max, err := strconv.Atoi(tx_max_str)
	if err != nil {
		msg := "updateboltcardwithpin: tx_max is not a valid integer"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	day_max_str := r.URL.Query().Get("day_max")
	day_max, err := strconv.Atoi(day_max_str)
	if err != nil {
		msg := "updateboltcardwithpin: day_max is not a valid integer"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	pin_enable_flag_str := r.URL.Query().Get("enable_pin")
	pin_enable_flag, err := strconv.ParseBool(pin_enable_flag_str)
	if err != nil {
		msg := "updateboltcardwithpin: enable_pin is not a valid boolean"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	pin_number := r.URL.Query().Get("pin_number")

	pin_limit_sats_str := r.URL.Query().Get("pin_limit_sats")
	pin_limit_sats, err := strconv.Atoi(pin_limit_sats_str)
	if err != nil {
		msg := "updateboltcardwithpin: pin_limit_sats is not a valid integer"
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
		msg := "updateboltcardwithpin: the card name does not exist in the database"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	// log the request

	log.WithFields(log.Fields{
		"card_name": card_name, "tx_max": tx_max, "day_max": day_max,
		"enable": enable_flag, "enable_pin": pin_enable_flag,
		"pin_number": pin_number, "pin_limit_sats": pin_limit_sats}).Info("updateboltcardwithpin API request")

	// update the card record

	if pin_number == "" {
		err = db.Update_card_with_part_pin(card_name, enable_flag, tx_max, day_max,
			pin_enable_flag, pin_limit_sats)
		if err != nil {
			log.Warn(err.Error())
			return
		}
	}

	if pin_number != "" {
		err = db.Update_card_with_pin(card_name, enable_flag, tx_max, day_max,
			pin_enable_flag, pin_number, pin_limit_sats)
		if err != nil {
			log.Warn(err.Error())
			return
		}
	}

	// send a response

	jsonData := []byte(`{"status":"OK"}`)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
