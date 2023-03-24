package lndhub

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/boltcard/boltcard/db"
	log "github.com/sirupsen/logrus"
)

type LndhubPayInvoiceRequest struct {
	Invoice    string `json:"invoice"`
	FreeAmount string `json:"freeamount"`
	LoginId    string `json:"loginid"`
}

func PayInvoice(cardPaymentId int, invoice string, amountSats int, loginId string, accessToken string) {

	lndhub_url := db.Get_setting("LNDHUB_URL")

	client := &http.Client{}

	//lndhub.payinvoice API call
	var payInvoiceRequest LndhubPayInvoiceRequest
	payInvoiceRequest.Invoice = invoice
	payInvoiceRequest.FreeAmount = strconv.Itoa(amountSats)
	payInvoiceRequest.LoginId = loginId

	req_payinvoice, err := json.Marshal(payInvoiceRequest)
	log.Info(string(req_payinvoice))
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": cardPaymentId}).Warn(err)
		return
	}

	req, err := http.NewRequest("POST", lndhub_url+"/payinvoice", bytes.NewBuffer(req_payinvoice))
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": cardPaymentId}).Warn(err)
		return
	}

	req.Header.Add("Access-Control-Allow-Origin", "*")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	res2, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": cardPaymentId}).Warn(err)
		return
	}

	defer res2.Body.Close()

	b2, err := io.ReadAll(res2.Body)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": cardPaymentId}).Warn(err)
		return
	}

	log.Info(string(b2))
}
