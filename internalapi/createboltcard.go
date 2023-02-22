package internalapi

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

func random_hex() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Warn(err.Error())
		return ""
	}

	return hex.EncodeToString(b)
}

func Createboltcard(w http.ResponseWriter, r *http.Request) {
	if db.Get_setting("FUNCTION_INTERNAL_API") != "ENABLE" {
		msg := "createboltcard: internal API function is not enabled"
		log.Debug(msg)
		resp_err.Write_message(w, msg)
		return
	}

	tx_max_str := r.URL.Query().Get("tx_max")
	tx_max, err := strconv.Atoi(tx_max_str)
	if err != nil {
		msg := "createboltcard: tx_max is not a valid integer"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	day_max_str := r.URL.Query().Get("day_max")
	day_max, err := strconv.Atoi(day_max_str)
	if err != nil {
		msg := "createboltcard: day_max is not a valid integer"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	enable_flag_str := r.URL.Query().Get("enable")
	enable_flag, err := strconv.ParseBool(enable_flag_str)
	if err != nil {
		msg := "createboltcard: enable is not a valid boolean"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	card_name := r.URL.Query().Get("card_name")

	uid_privacy_flag_str := r.URL.Query().Get("uid_privacy")
	uid_privacy_flag, err := strconv.ParseBool(uid_privacy_flag_str)
	if err != nil {
		msg := "createboltcard: uid_privacy is not a valid boolean"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	allow_neg_bal_flag_str := r.URL.Query().Get("allow_neg_bal")
	allow_neg_bal_flag, err := strconv.ParseBool(allow_neg_bal_flag_str)
	if err != nil {
		msg := "createboltcard: allow_neg_bal is not a valid boolean"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	// check if card_name already exists

	card_count, err := db.Get_card_name_count(card_name)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	if card_count > 0 {
		msg := "createboltcard: the card name already exists in the database"
		log.Warn(msg)
		resp_err.Write_message(w, msg)
		return
	}

	// log the request

	log.WithFields(log.Fields{
		"card_name": card_name, "tx_max": tx_max, "day_max": day_max,
		"enable": enable_flag, "uid_privacy": uid_privacy_flag,
		"allow_neg_bal": allow_neg_bal_flag}).Info("createboltcard API request")

	// create the keys

	one_time_code := random_hex()
	k0_auth_key := random_hex()
	k2_cmac_key := random_hex()
	k3 := random_hex()
	k4 := random_hex()

	// create the new card record

	err = db.Insert_card(one_time_code, k0_auth_key, k2_cmac_key, k3, k4,
		tx_max, day_max, enable_flag, card_name,
		uid_privacy_flag, allow_neg_bal_flag)
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
		"card_name": card_name, "url": url}).Info("createboltcard API response")

	jsonData := []byte(`{"status":"OK",` +
		`"url":"` + url + `"}`)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
