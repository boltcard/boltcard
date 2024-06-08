package lnd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"strconv"
	"time"

	decodepay "github.com/fiatjaf/ln-decodepay"
	lnrpc "github.com/lightningnetwork/lnd/lnrpc"
	invoicesrpc "github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	routerrpc "github.com/lightningnetwork/lnd/lnrpc/routerrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"gopkg.in/macaroon.v2"

	"github.com/boltcard/boltcard/db"
	"github.com/boltcard/boltcard/email"
)

type rpcCreds map[string]string

func (m rpcCreds) RequireTransportSecurity() bool { return true }
func (m rpcCreds) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return m, nil
}
func newCreds(bytes []byte) rpcCreds {
	creds := make(map[string]string)
	creds["macaroon"] = hex.EncodeToString(bytes)
	return creds
}

func getGrpcConn(hostname string, port int, tlsFile, macaroonFile string) *grpc.ClientConn {
	macaroonBytes, err := ioutil.ReadFile(macaroonFile)
	if err != nil {
		log.Println("Cannot read macaroon file .. ", err)
		panic(err)
	}

	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
		log.Println("Cannot unmarshal macaroon .. ", err)
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transportCredentials, err := credentials.NewClientTLSFromFile(tlsFile, hostname)
	if err != nil {
		panic(err)
	}

	fullHostname := fmt.Sprintf("%s:%d", hostname, port)

	connection, err := grpc.DialContext(ctx, fullHostname, []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(transportCredentials),
		grpc.WithPerRPCCredentials(newCreds(macaroonBytes)),
	}...)
	if err != nil {
		log.Printf("unable to connect to %s", fullHostname)
		panic(err)
	}

	return connection
}

// https://api.lightning.community/?shell#addinvoice

func Add_invoice(amount_sat int64, metadata string) (payment_request string, r_hash []byte, return_err error) {

	ln_port, err := strconv.Atoi(db.Get_setting("LN_PORT"))
	if err != nil {
		return "", nil, err
	}
	ln_invoice_expiry, err := strconv.ParseInt(db.Get_setting("LN_INVOICE_EXPIRY_SEC"), 10, 64)
	if err != nil {
		return "", nil, err
	}

	dh := sha256.Sum256([]byte(metadata))

	connection := getGrpcConn(
		db.Get_setting("LN_HOST"),
		ln_port,
		db.Get_setting("LN_TLS_FILE"),
		db.Get_setting("LN_MACAROON_FILE"))

	l_client := lnrpc.NewLightningClient(connection)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := l_client.AddInvoice(ctx, &lnrpc.Invoice{
		Value:           amount_sat,
		DescriptionHash: dh[:],
		Expiry:          ln_invoice_expiry,
	})

	if err != nil {
		return "", nil, err
	}

	return result.PaymentRequest, result.RHash, nil
}

// https://api.lightning.community/?shell#subscribesingleinvoice

func Monitor_invoice_state(r_hash []byte) {

	// SubscribeSingleInvoice

	// get node parameters from environment variables

	ln_port, err := strconv.Atoi(db.Get_setting("LN_PORT"))
	if err != nil {
		log.Warn(err)
		return
	}
	ln_invoice_expiry, err := strconv.Atoi(db.Get_setting("LN_INVOICE_EXPIRY_SEC"))
	if err != nil {
		log.Warn(err)
		return
	}

	connection := getGrpcConn(
		db.Get_setting("LN_HOST"),
		ln_port,
		db.Get_setting("LN_TLS_FILE"),
		db.Get_setting("LN_MACAROON_FILE"))

	i_client := invoicesrpc.NewInvoicesClient(connection)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(ln_invoice_expiry)*time.Second)
	defer cancel()

	stream, err := i_client.SubscribeSingleInvoice(ctx, &invoicesrpc.SubscribeSingleInvoiceRequest{
		RHash: r_hash})
	if err != nil {
		log.WithFields(log.Fields{"r_hash": hex.EncodeToString(r_hash)}).Warn(err)
		return
	}

	for {
		update, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.WithFields(log.Fields{"r_hash": hex.EncodeToString(r_hash)}).Warn(err)
			return
		}

		invoice_state := lnrpc.Invoice_InvoiceState_name[int32(update.State)]

		log.WithFields(
			log.Fields{
				"r_hash":        hex.EncodeToString(r_hash),
				"invoice_state": invoice_state,
			}).Info("invoice state updated")

		db.Update_receipt_state(hex.EncodeToString(r_hash), invoice_state)
	}

	connection.Close()

	// send email

	card_id, err := db.Get_card_id_for_r_hash(hex.EncodeToString(r_hash))
	if err != nil {
		log.WithFields(log.Fields{"r_hash": hex.EncodeToString(r_hash)}).Warn(err)
		return
	}

	log.WithFields(log.Fields{"r_hash": hex.EncodeToString(r_hash), "card_id": card_id}).Debug("card found")

	c, err := db.Get_card_from_card_id(card_id)
	if err != nil {
		log.WithFields(log.Fields{"r_hash": hex.EncodeToString(r_hash)}).Warn(err)
		return
	}

	if c.Email_enable != "Y" {
		log.Debug("email is not enabled for the card")
		return
	}

	go email.Send_balance_email(c.Email_address, card_id)

	return
}

// https://api.lightning.community/?shell#sendpaymentv2

func PayInvoice(card_payment_id int, invoice string) {

	// SendPaymentV2

	// get node parameters from environment variables

	ln_port, err := strconv.Atoi(db.Get_setting("LN_PORT"))
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": card_payment_id}).Warn(err)
		return
	}

	connection := getGrpcConn(
		db.Get_setting("LN_HOST"),
		ln_port,
		db.Get_setting("LN_TLS_FILE"),
		db.Get_setting("LN_MACAROON_FILE"))

	r_client := routerrpc.NewRouterClient(connection)

	fee_limit_sat_str := db.Get_setting("FEE_LIMIT_SAT")
	fee_limit_sat, err := strconv.ParseInt(fee_limit_sat_str, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": card_payment_id}).Warn(err)
		return
	}

	fee_limit_percent_str := db.Get_setting("FEE_LIMIT_PERCENT")
	fee_limit_percent, err := strconv.ParseFloat(fee_limit_percent_str, 64)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": card_payment_id}).Warn(err)
		return
	}

	bolt11, _ := decodepay.Decodepay(invoice)
	invoice_msats := bolt11.MSatoshi
	invoice_expiry := bolt11.Expiry

	fee_limit_product := int64((fee_limit_percent / 100) * (float64(invoice_msats) / 1000))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(invoice_expiry)*time.Second)
	defer cancel()

	stream, err := r_client.SendPaymentV2(ctx, &routerrpc.SendPaymentRequest{
		PaymentRequest:    invoice,
		NoInflightUpdates: true,
		TimeoutSeconds:    30,
		FeeLimitSat:       fee_limit_sat + fee_limit_product})

	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": card_payment_id}).Warn(err)
		return
	}

	for {
		update, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.WithFields(log.Fields{"card_payment_id": card_payment_id}).Warn(err)
			return
		}

		payment_status := lnrpc.Payment_PaymentStatus_name[int32(update.Status)]
		failure_reason := lnrpc.PaymentFailureReason_name[int32(update.FailureReason)]

		log.WithFields(log.Fields{"card_payment_id": card_payment_id}).Info("payment failure reason : ", failure_reason)
		log.WithFields(log.Fields{"card_payment_id": card_payment_id}).Info("payment status : ", payment_status)

		err = db.Update_payment_status(card_payment_id, payment_status, failure_reason)
		if err != nil {
			log.WithFields(log.Fields{"card_payment_id": card_payment_id}).Warn(err)
			return
		}
	}

	connection.Close()

	// send email

	card_id, err := db.Get_card_id_for_card_payment_id(card_payment_id)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": card_payment_id}).Warn(err)
		return
	}

	log.WithFields(log.Fields{"card_payment_id": card_payment_id, "card_id": card_id}).Debug("card found")

	c, err := db.Get_card_from_card_id(card_id)
	if err != nil {
		log.WithFields(log.Fields{"card_payment_id": card_payment_id}).Warn(err)
		return
	}

	if c.Email_enable != "Y" {
		log.Debug("email is not enabled for the card")
		return
	}

	go email.Send_balance_email(c.Email_address, card_id)

	return
}
