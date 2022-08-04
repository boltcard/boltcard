package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

type Response struct {
	Tag                string `json:"tag"`
	Callback           string `json:"callback"`
	K1                 string `json:"k1"`
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

func parse_request(req *http.Request) (int, error) {

	url := req.URL.RequestURI()
	log.Debug("ln url: ", url)

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

	// get card record from database for UID

	uid_str := hex.EncodeToString(uid)
	log.Debug("card UID: ", uid_str)

	c, err := db_get_card_from_uid(uid_str)

	if err != nil {
		return 0, errors.New("card not found for UID")
	}

	// check if card is enabled

	if c.enable_flag != "Y" {
		return 0, errors.New("card enable is not set to Y")
	}

	// check cmac

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

	key_sdm_file_read_mac, err := hex.DecodeString(c.aes_cmac)
	if err != nil {
		return 0, err
	}

	cmac_verified, err := crypto_aes_cmac(key_sdm_file_read_mac, sv2, ba_c)
	if err != nil {
		return 0, err
	}

	if cmac_verified == false {
		return 0, errors.New("CMAC incorrect")
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

	card_id, err := parse_request(req)

	if err != nil {
		log.Debug(err.Error())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		jsonData := []byte(`{"status":"ERROR","reason":"bad request"}`)
		w.Write(jsonData)

		return
	}

	k1, err := create_k1()
	if err != nil {
		log.Warn(err.Error())
		return
	}

	// store k1 in database and include in response

	err = db_insert_payment(card_id, k1)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	host_domain := os.Getenv("HOST_DOMAIN")
	lnurlw_cb_url := "https://" + host_domain + "/cb"

	min_withdraw_sats_str := os.Getenv("MIN_WITHDRAW_SATS")
	min_withdraw_sats, err := strconv.Atoi(min_withdraw_sats_str)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	max_withdraw_sats_str := os.Getenv("MAX_WITHDRAW_SATS")
	max_withdraw_sats, err := strconv.Atoi(max_withdraw_sats_str)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	response := Response{}
	response.Tag = "withdrawRequest"
	response.Callback = lnurlw_cb_url
	response.K1 = k1
	response.DefaultDescription = "WWT withdrawal"
	response.MinWithdrawable = min_withdraw_sats * 1000 // milliSats
	response.MaxWithdrawable = max_withdraw_sats * 1000 // milliSats

	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Warn(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
