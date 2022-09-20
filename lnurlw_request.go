package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Response struct {
	Tag                string `json:"tag"`
	Callback           string `json:"callback"`
	LnurlwK1           string `json:"k1"`
	DefaultDescription string `json:"defaultDescription"`
	MinWithdrawable    int    `json:"minWithdrawable"`
	MaxWithdrawable    int    `json:"maxWithdrawable"`
}

func get_p_c(req *http.Request, p_name string, c_name string) (p string, c string) {

	params_p, ok := req.URL.Query()[p_name]

	if !ok || len(params_p[0]) < 1 {
		return "", ""
	}

	params_c, ok := req.URL.Query()[c_name]

	if !ok || len(params_c[0]) < 1 {
		return "", ""
	}

	p = params_p[0]
	c = params_c[0]

	return
}

func check_cmac(uid []byte, ctr []byte, k2_cmac_key []byte, cmac []byte) (bool, error) {

	sv2 := make([]byte, 16)
	sv2[0] = 0x3c
	sv2[1] = 0xc3
	sv2[2] = 0x00
	sv2[3] = 0x01
	sv2[4] = 0x00
	sv2[5] = 0x80
	sv2[6] = uid[0]
	sv2[7] = uid[1]
	sv2[8] = uid[2]
	sv2[9] = uid[3]
	sv2[10] = uid[4]
	sv2[11] = uid[5]
	sv2[12] = uid[6]
	sv2[13] = ctr[0]
	sv2[14] = ctr[1]
	sv2[15] = ctr[2]

	cmac_verified, err := crypto_aes_cmac(k2_cmac_key, sv2, cmac)

	if err != nil {
		return false, err
	}

	return cmac_verified, nil
}

func setup_card_record(uid string, ctr uint32, uid_bin []byte, ctr_bin []byte, cmac []byte) error {

	// find the card record by matching the cmac
	//  get possible card records from the database

	cards, err := db_get_cards_blank_uid()

	if err != nil {
		return errors.New("db_get_cards_blank_uid errored")
	}

	//  check card records for a matching cmac

	for i, card := range cards {
		// check the cmac

		k2_cmac_key, err := hex.DecodeString(card.k2_cmac_key)

		if err != nil {
			return errors.New("card.k2_cmac_key decode failed")
		}

		cmac_valid, err := check_cmac(uid_bin, ctr_bin, k2_cmac_key, cmac)

		if err != nil {
			return err
		}

		if cmac_valid == true {
			log.WithFields(log.Fields{
				"i":                i,
				"card.card_id":     card.card_id,
				"card.k2_cmac_key": card.k2_cmac_key,
			}).Info("cmac match found")

			// store the uid and ctr in the card record
			err := db_update_card_uid_ctr(card.card_id, uid, ctr)

			if err != nil {
				return err
			}

			return nil
		}
	}

	log.Info("card record not found")

	return nil
}

func parse_request(req *http.Request) (int, error) {

	pid := os.Getpid()
	url := req.URL.RequestURI()
	log.WithFields(log.Fields{"pid": pid, "url": url}).Debug("ln request")

	param_p, param_c := get_p_c(req, "p", "c")

	ba_p, err := hex.DecodeString(param_p)

	if err != nil {
		return 0, errors.New("p parameter not valid hex")
	}

	ba_c, err := hex.DecodeString(param_c)

	if err != nil {
		return 0, errors.New("c parameter not valid hex")
	}

	if len(ba_p) != 16 {
		return 0, errors.New("p parameter length not valid")
	}

	if len(ba_c) != 8 {
		return 0, errors.New("c parameter length not valid")
	}

	// decrypt p with aes_decrypt_key

	aes_decrypt_key := os.Getenv("AES_DECRYPT_KEY")

	key_sdm_file_read, err := hex.DecodeString(aes_decrypt_key)

	if err != nil {
		return 0, err
	}

	dec_p, err := crypto_aes_decrypt(key_sdm_file_read, ba_p)

	if err != nil {
		return 0, err
	}

	if dec_p[0] != 0xC7 {
		return 0, errors.New("decrypted data not starting with 0xC7")
	}

	uid := dec_p[1:8]
	ctr := dec_p[8:11]

	ctr_int := uint32(ctr[2])<<16 | uint32(ctr[1])<<8 | uint32(ctr[0])

	// set up uid & ctr for card record if needed

	uid_str := hex.EncodeToString(uid)

	log.WithFields(log.Fields{"uid": uid_str, "ctr": ctr_int}).Info("decrypted card data")

	card_count, err := db_get_card_count_for_uid(uid_str)

	if err != nil {
		return 0, errors.New("could not get card count for uid")
	}

	if card_count == 0 {
		setup_card_record(uid_str, ctr_int, uid, ctr, ba_c)
	}

	if card_count > 1 {
		return 0, errors.New("more than one card found for uid")
	}

	// check card payment rules and make payment if appropriate

	// get card record from database for UID

	c, err := db_get_card_from_uid(uid_str)

	if err != nil {
		return 0, errors.New("card not found for uid")
	}

	// check if card is enabled

	if c.lnurlw_enable != "Y" {
		return 0, errors.New("card lnurlw enable is not set to Y")
	}

	// check cmac

	k2_cmac_key, err := hex.DecodeString(c.k2_cmac_key)

	if err != nil {
		return 0, err
	}

	cmac_valid, err := check_cmac(uid, ctr, k2_cmac_key, ba_c)

	if err != nil {
		return 0, err
	}

	if cmac_valid == false {
		return 0, errors.New("cmac incorrect")
	}

	// check and update last_counter_value

	counter_ok, err := db_check_and_update_counter(c.card_id, ctr_int)

	if err != nil {
		return 0, err
	}

	if counter_ok == false {
		return 0, errors.New("counter not increasing")
	}

	log.WithFields(log.Fields{"card_id": c.card_id, "counter": ctr_int}).Info("validated")

	return c.card_id, nil
}

func lnurlw_response(w http.ResponseWriter, req *http.Request) {

	env_host_domain := os.Getenv("HOST_DOMAIN")
	if req.Host != env_host_domain {
		log.Warn("wrong host domain")
		write_error(w)
		return
	}

	card_id, err := parse_request(req)

	if err != nil {
		log.Debug(err.Error())
		write_error(w)
		return
	}

	lnurlw_k1, err := create_k1()

	if err != nil {
		log.Warn(err.Error())
		write_error(w)
		return
	}

	// store k1 in database and include in response

	err = db_insert_payment(card_id, lnurlw_k1)

	if err != nil {
		log.Warn(err.Error())
		write_error(w)
		return
	}

	lnurlw_cb_url := ""
	if strings.HasSuffix(req.Host, ".onion") {
		lnurlw_cb_url = "http://" + req.Host + "/cb"
	} else {
		lnurlw_cb_url = "https://" + req.Host + "/cb"
	}

	min_withdraw_sats_str := os.Getenv("MIN_WITHDRAW_SATS")
	min_withdraw_sats, err := strconv.Atoi(min_withdraw_sats_str)

	if err != nil {
		log.Warn(err.Error())
		write_error(w)
		return
	}

	max_withdraw_sats_str := os.Getenv("MAX_WITHDRAW_SATS")
	max_withdraw_sats, err := strconv.Atoi(max_withdraw_sats_str)

	if err != nil {
		log.Warn(err.Error())
		write_error(w)
		return
	}

	response := Response{}
	response.Tag = "withdrawRequest"
	response.Callback = lnurlw_cb_url
	response.LnurlwK1 = lnurlw_k1
	response.DefaultDescription = "WWT withdrawal"
	response.MinWithdrawable = min_withdraw_sats * 1000 // milliSats
	response.MaxWithdrawable = max_withdraw_sats * 1000 // milliSats

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
