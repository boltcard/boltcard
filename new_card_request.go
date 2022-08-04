package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type NewCardResponse struct {
	K0                 string `json:"k0"`
	K1                 string `json:"k1"`
	K2                 string `json:"k2"`
}

func new_card_request(w http.ResponseWriter, req *http.Request) {

	url := req.URL.RequestURI()
	log.Debug("new_card url: ", url)

	params_a, ok := req.URL.Query()["a"]
	if !ok || len(params_a[0]) < 1 {
		log.Debug("a not found")
		return
	}

	a := params_a[0]

	if a == "00000000000000000000000000000000" {
		response := NewCardResponse{}
		response.K0 = "11111111111111111111111111111111"
		response.K1 = "22222222222222222222222222222222"
		response.K2 = "33333333333333333333333333333333"
		log.Debug("special a = 0...0")

		jsonData, err := json.Marshal(response)
		if err != nil {
			log.Warn(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)

		return;
	}

	c, err := db_get_new_card(a)
	if err != nil {
		log.Warn(err)
		return
	}

	aes_decrypt_key := os.Getenv("AES_DECRYPT_KEY")

	response := NewCardResponse{}
	response.K0 = c.lock_key
	response.K1 = aes_decrypt_key
	response.K2 = c.aes_cmac

	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Warn(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
