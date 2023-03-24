package lnurlw

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/boltcard/boltcard/db"
	"github.com/boltcard/boltcard/lnd"
	"github.com/boltcard/boltcard/resp_err"
	decodepay "github.com/fiatjaf/ln-decodepay"
	log "github.com/sirupsen/logrus"
)

type LndhubAuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LndhubAuthResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type LndhubPayInvoiceRequest struct {
	Invoice    string `json:"invoice"`
	FreeAmount string `json:"freeamount"`
	LoginId    string `json:"loginid"`
}

func lndhub_payment(w http.ResponseWriter, p *db.Payment, bolt11 decodepay.Bolt11, param_pr string) {

	//get setting for LNDHUB_URL
	lndhub_url := db.Get_setting("LNDHUB_URL")

	//get lndhub login details from database
	c, err := db.Get_card_from_card_id(p.Card_id)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	//lndhub.auth API call
	//the login JSON is held in the Card_name field
	// as "login:password"
	card_name_parts := strings.Split(c.Card_name, ":")

	if len(card_name_parts) != 2 {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	if len(card_name_parts[0]) != 20 || len(card_name_parts[1]) != 20 {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	var lhAuthRequest LndhubAuthRequest
	lhAuthRequest.Login = card_name_parts[0]
	lhAuthRequest.Password = card_name_parts[1]

	authReq, err := json.Marshal(lhAuthRequest)
	log.Info(string(authReq))

	req_auth, err := http.NewRequest("POST", lndhub_url+"/auth", bytes.NewBuffer(authReq))
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	req_auth.Header.Add("Access-Control-Allow-Origin", "*")
	req_auth.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp_auth, err := client.Do(req_auth)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	defer resp_auth.Body.Close()

	resp_auth_bytes, err := io.ReadAll(resp_auth.Body)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	var auth_keys LndhubAuthResponse

	err = json.Unmarshal([]byte(resp_auth_bytes), &auth_keys)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	//lndhub.payinvoice API call
	var payInvoiceRequest LndhubPayInvoiceRequest
	payInvoiceRequest.Invoice = param_pr
	payInvoiceRequest.FreeAmount = strconv.Itoa(int(bolt11.MSatoshi / 1000))
	payInvoiceRequest.LoginId = card_name_parts[0]

	req_payinvoice, err := json.Marshal(payInvoiceRequest)
	log.Info(string(req_payinvoice))
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	req, err := http.NewRequest("POST", lndhub_url+"/payinvoice", bytes.NewBuffer(req_payinvoice))
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	req.Header.Add("Access-Control-Allow-Origin", "*")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+auth_keys.AccessToken)

	res2, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	defer res2.Body.Close()

	b2, err := io.ReadAll(res2.Body)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	log.Info(string(b2))

	log.Debug("sending 'status OK' response")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(`{"status":"OK"}`)
	w.Write(jsonData)

}

func lnd_payment(w http.ResponseWriter, p *db.Payment, bolt11 decodepay.Bolt11, param_pr string) {

	// check amount limits
	invoice_sats := int(bolt11.MSatoshi / 1000)

	day_total_sats, err := db.Get_card_totals(p.Card_id)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	c, err := db.Get_card_from_card_id(p.Card_id)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	if invoice_sats > c.Tx_limit_sats {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("invoice_sats: ", invoice_sats)
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("tx_limit_sats: ", c.Tx_limit_sats)
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("over tx_limit_sats!")
		resp_err.Write(w)
		return
	}

	if day_total_sats+invoice_sats > c.Day_limit_sats {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("invoice_sats: ", invoice_sats)
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("day_total_sats: ", day_total_sats)
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("day_limit_sats: ", c.Day_limit_sats)
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("over day_limit_sats!")
		resp_err.Write(w)
		return
	}

	// check the card balance if marked as 'must stay above zero' (default)
	//  i.e. cards.allow_negative_balance == 'N'
	if c.Allow_negative_balance != "Y" {
		card_total, err := db.Get_card_total_sats(p.Card_id)
		if err != nil {
			log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
			resp_err.Write(w)
			return
		}

		if card_total-invoice_sats < 0 {
			log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn("not enough balance")
			resp_err.Write(w)
			return
		}
	}

	log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("paying invoice")

	// update paid_flag so we only attempt payment once
	err = db.Update_payment_paid(p.Card_payment_id)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	// https://github.com/fiatjaf/lnurl-rfc/blob/luds/03.md
	//
	// LN SERVICE sends a {"status": "OK"} or
	// {"status": "ERROR", "reason": "error details..."}
	//  JSON response and then attempts to pay the invoices asynchronously.

	go lnd.Pay_invoice(p.Card_payment_id, param_pr)

	log.Debug("sending 'status OK' response")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(`{"status":"OK"}`)
	w.Write(jsonData)
}

func Callback(w http.ResponseWriter, req *http.Request) {

	env_host_domain := db.Get_setting("HOST_DOMAIN")
	if req.Host != env_host_domain {
		log.Warn("wrong host domain")
		resp_err.Write(w)
		return
	}

	url := req.URL.RequestURI()
	log.WithFields(log.Fields{"url": url}).Debug("cb request")

	// check k1 value
	params_k1, ok := req.URL.Query()["k1"]

	if !ok || len(params_k1[0]) < 1 {
		log.WithFields(log.Fields{"url": url}).Debug("k1 not found")
		resp_err.Write(w)
		return
	}

	param_k1 := params_k1[0]

	p, err := db.Get_payment_k1(param_k1)
	if err != nil {
		log.WithFields(log.Fields{"url": url, "k1": param_k1}).Warn(err)
		resp_err.Write(w)
		return
	}

	// check that payment has not been made
	if p.Paid_flag != "N" {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("payment already made")
		resp_err.Write(w)
		return
	}

	// check if lnurlw_request has timed out
	lnurlw_timeout, err := db.Check_lnurlw_timeout(p.Card_payment_id)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}
	if lnurlw_timeout == true {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("lnurlw request has timed out")
		resp_err.Write(w)
		return
	}

	params_pr, ok := req.URL.Query()["pr"]
	if !ok || len(params_pr[0]) < 1 {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn("pr field not found")
		resp_err.Write(w)
		return
	}

	param_pr := params_pr[0]
	bolt11, _ := decodepay.Decodepay(param_pr)

	// record the lightning invoice
	err = db.Update_payment_invoice(p.Card_payment_id, param_pr, bolt11.MSatoshi)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Debug("checking payment rules")

	// check if we are only sending funds to a defined test node
	testnode := db.Get_setting("LN_TESTNODE")
	if testnode != "" && bolt11.Payee != testnode {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("rejected as not the defined test node")
		resp_err.Write(w)
		return
	}

	//check if we are using LND or LNDHUB for payment
	lndhub := db.Get_setting("FUNCTION_LNDHUB")
	if lndhub == "ENABLE" {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("initiating lndhub payment")
		lndhub_payment(w, p, bolt11, param_pr)
	} else {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("initiating lnd payment")
		lnd_payment(w, p, bolt11, param_pr)
	}
}
