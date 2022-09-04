package main

import (
	"net/http"
	"os"

	decodepay "github.com/fiatjaf/ln-decodepay"
	log "github.com/sirupsen/logrus"
)

func lnurlw_callback(w http.ResponseWriter, req *http.Request) {

	url := req.URL.RequestURI()
	log.WithFields(log.Fields{"url": url}).Debug("cb request")

	// check k1 value
	params_k1, ok := req.URL.Query()["k1"]

	if !ok || len(params_k1[0]) < 1 {
		log.WithFields(log.Fields{"url": url}).Debug("k1 not found")
		write_error(w)
		return
	}

	param_k1 := params_k1[0]

	p, err := db_get_payment_k1(param_k1)
	if err != nil {
		log.WithFields(log.Fields{"url": url, "k1": param_k1}).Warn(err)
		write_error(w)
		return
	}

	// check that payment has not been made
	if p.paid_flag != "N" {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("payment already made")
		write_error(w)
		return
	}

	// check if lnurlw_request has timed out
	lnurlw_timeout, err := db_check_lnurlw_timeout(p.card_payment_id)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Warn(err)
		write_error(w)
		return
	}
	if lnurlw_timeout {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("lnurlw request has timed out")
		write_error(w)
		return
	}

	params_pr, ok := req.URL.Query()["pr"]
	if !ok || len(params_pr[0]) < 1 {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Warn("pr field not found")
		write_error(w)
		return
	}

	param_pr := params_pr[0]
	bolt11, _ := decodepay.Decodepay(param_pr)

	// record the lightning invoice
	err = db_update_payment_invoice(p.card_payment_id, param_pr, bolt11.MSatoshi)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Warn(err)
		write_error(w)
		return
	}

	log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Debug("checking payment rules")

	// check if we are only sending funds to a defined test node
	testnode := os.Getenv("LN_TESTNODE")
	if testnode != "" && bolt11.Payee != testnode {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("rejected as not the defined test node")
		write_error(w)
		return
	}

	// check amount limits

	invoice_sats := int(bolt11.MSatoshi / 1000)

	day_total_sats, err := db_get_card_totals(p.card_id)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Warn(err)
		write_error(w)
		return
	}

	c, err := db_get_card_from_card_id(p.card_id)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Warn(err)
		write_error(w)
		return
	}

	if invoice_sats > c.tx_limit_sats {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("invoice_sats: ", invoice_sats)
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("tx_limit_sats: ", c.tx_limit_sats)
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("over tx_limit_sats!")
		write_error(w)
		return
	}

	if day_total_sats+invoice_sats > c.day_limit_sats {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("invoice_sats: ", invoice_sats)
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("day_total_sats: ", day_total_sats)
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("day_limit_sats: ", c.day_limit_sats)
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("over day_limit_sats!")
		write_error(w)
		return
	}

	log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("paying invoice")

	// update paid_flag so we only attempt payment once
	err = db_update_payment_paid(p.card_payment_id)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Warn(err)
		write_error(w)
		return
	}

	payment_status, failure_reason, err := pay_invoice(param_pr)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Warn(err)
		write_error(w)
		return
	}

	if failure_reason != "FAILURE_REASON_NONE" {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("payment failure reason : ", failure_reason)
		write_error(w)
	}

	log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Info("payment status : ", payment_status)

	// store result in database
	err = db_update_payment_status(p.card_payment_id, payment_status, failure_reason)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": p.card_payment_id}).Warn(err)
		write_error(w)
		return
	}

	// https://github.com/fiatjaf/lnurl-rfc/blob/luds/03.md
	//
	// LN SERVICE sends a {"status": "OK"} or
	// {"status": "ERROR", "reason": "error details..."}
	//  JSON response and then attempts to pay the invoices asynchronously.

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData := []byte(`{"status":"OK"}`)
	w.Write(jsonData)
}
