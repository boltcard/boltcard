package main

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/boltcard/boltcard/db"
	"github.com/boltcard/boltcard/resp_err"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

func updateboltcard(w http.ResponseWriter, r *http.Request) {
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

	// check if card_name already exists
//TODO: allow multiple deactivated cards with the same card_name
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

	// create the keys

	one_time_code := random_hex()
	k0_auth_key := random_hex()
	k2_cmac_key := random_hex()
	k3 := random_hex()
	k4 := random_hex()

	// update the card record

	err = db.Update_card(card_name, tx_max, enable_flag)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	// return the URI + one_time_code

	hostdomain := db.Get_setting("HOST_DOMAIN")
	url := ""
	if strings.HasSuffix(hostdomain, ".onion") {
		url = "http://" + hostdomain + "/new?a=" + one_time_code
	} else {
		url = "https://" + hostdomain + "/new?a=" + one_time_code
	}

	// log the response

	log.WithFields(log.Fields{
		"card_name": card_name, "url": url}).Info("updateboltcard API response")

	jsonData := []byte(`{"status":"OK",` +
		`"url":"` + url + `"}`)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
