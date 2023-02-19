package lnurlw

import (
	"bytes"
	"github.com/boltcard/boltcard/db"
	"github.com/boltcard/boltcard/lnd"
	"github.com/boltcard/boltcard/resp_err"
	decodepay "github.com/fiatjaf/ln-decodepay"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func lndhub_payment(w http.ResponseWriter, p *db.Payment) {

	//get setting for LNDHUB_URL
	lndhub_url := "https://" + db.Get_setting("LNDHUB_URL")

	//get lndhub login details from database
	c, err := db.Get_card_from_card_id(p.Card_id)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	//lndhub.auth API call
	//the login JSON is held in the Card_name field
	body := []byte(c.Card_name)

	r, err := http.NewRequest("POST", lndhub_url+"/auth", bytes.NewBuffer(body))
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	r.Header.Add("Access-Control-Allow-Origin", "*")
	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	defer res.Body.Close()

	b, err2 := io.ReadAll(res.Body)
	if err2 != nil {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Warn(err)
		resp_err.Write(w)
		return
	}

	log.Info(string(b))

	//	fmt.Println(string(b))

	//lndhub.payinvoice API call
}

func lnd_payment(w http.ResponseWriter, bolt11 decodepay.Bolt11, p *db.Payment, param_pr string) {

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
		lndhub_payment(w, p)
	} else {
		log.WithFields(log.Fields{"card_payment_id": p.Card_payment_id}).Info("initiating lnd payment")
		lnd_payment(w, bolt11, p, param_pr)
	}
}
