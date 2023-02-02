package main

import (
	log "github.com/sirupsen/logrus"
	"strconv"
	"net/http"
)

type card_wipe_info struct {
	id  int
	k0  string
	k1  string
	k2  string
	k3  string
	k4  string
	uid string
}

func wipeboltcard(w http.ResponseWriter, r *http.Request) {
        if db_get_setting("FUNCTION_INTERNAL_API") != "ENABLE" {
                msg := "wipeboltcard: internal API function is not enabled"
                log.Debug(msg)
                write_error_message(w, msg)
                return
        }

	card_name := r.URL.Query().Get("card_name")

	// check if card_name has been given

	if card_name == "" {
                msg := "wipeboltcard: the card name must be set"
                log.Warn(msg)
                write_error_message(w, msg)
                return
	}

	// check if card_name exists

	card_count, err := db_get_card_name_count(card_name)

        if card_count == 0 {
                msg := "the card name does not exist in the database"
                log.Warn(msg)
                write_error_message(w, msg)
                return
        }

	// set the card as wiped and disabled, get the keys

	card_wipe_info_values, err := db_wipe_card(card_name)
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
		`"id": ` + strconv.Itoa(card_wipe_info_values.id) + `,` +
		`"k0": "` + card_wipe_info_values.k0 + `",` +
		`"k1": "` + card_wipe_info_values.k1 + `",` +
		`"k2": "` + card_wipe_info_values.k2 + `",` +
		`"k3": "` + card_wipe_info_values.k3 + `",` +
		`"k4": "` + card_wipe_info_values.k4 + `",` +
		`"uid": "` + card_wipe_info_values.uid + `",` +
		`"version": 1"}`

        // log the response

        log.WithFields(log.Fields{
                "card_name": card_name, "response": jsonData}).Info("wipeboltcard API response")

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(jsonData))
}
