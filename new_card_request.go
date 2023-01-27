package main

import (
	"database/sql"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
)

/**
 * @api {get} /new/:a Request information to create a new bolt card
 * @apiName NewBoltCard
 * @apiGroup BoltCardService
 *
 * @apiParam {String} a one time authentication code
 *
 * @apiSuccess {String} protocol_name name of the protocol message
 * @apiSuccess {Int} protocol_version version of the protocol message
 * @apiSuccess {String} card_name user friendly card name
 * @apiSuccess {String} lnurlw_base base for creating the lnurlw on the card
 * @apiSuccess {String} k0 Key 0 - authorisation key
 * @apiSuccess {String} k1 Key 1 - decryption key
 * @apiSuccess {String} k2 Key 2 - authentication key
 * @apiSuccess {String} k3 Key 3 - NXP documents say this must be set
 * @apiSuccess {String} k4 Key 4 - NXP documents say this must be set
 * @apiSuccess {String} uid_privacy - set up the card for the UID to be private
 */

type NewCardResponse struct {
	PROTOCOL_NAME    string `json:"protocol_name"`
	PROTOCOL_VERSION int    `json:"protocol_version"`
	CARD_NAME        string `json:"card_name"`
	LNURLW_BASE      string `json:"lnurlw_base"`
	K0               string `json:"k0"`
	K1               string `json:"k1"`
	K2               string `json:"k2"`
	K3               string `json:"k3"`
	K4               string `json:"k4"`
	UID_PRIVACY	 string `json:"uid_privacy"`
}

func new_card_request(w http.ResponseWriter, req *http.Request) {

	url := req.URL.RequestURI()
	log.Debug("new_card url: ", url)

	params_a, ok := req.URL.Query()["a"]
	if !ok || len(params_a[0]) < 1 {
		log.Debug("a not found")
		write_error(w)
		return
	}

	a := params_a[0]

	lnurlw_base := "lnurlw://" + req.Host + "/ln"

	c, err := db_get_new_card(a)

	if err == sql.ErrNoRows {
		log.Debug(err)
		write_error_message(w, "one time code was used or card was wiped or card does not exist")
		return
	}

	if err != nil {
		log.Warn(err)
		write_error(w)
		return
	}

	k1_decrypt_key := db_get_setting("AES_DECRYPT_KEY")

	response := NewCardResponse{}
	response.PROTOCOL_NAME = "create_bolt_card_response"
	response.PROTOCOL_VERSION = 2
	response.CARD_NAME = c.card_name
	response.LNURLW_BASE = lnurlw_base
	response.K0 = c.k0_auth_key
	response.K1 = k1_decrypt_key
	response.K2 = c.k2_cmac_key
	response.K3 = c.k3
	response.K4 = c.k4
	response.UID_PRIVACY = c.uid_privacy

	log.SetFormatter(&log.JSONFormatter{
		DisableHTMLEscape: true,
	})
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Warn(err)
		write_error(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
