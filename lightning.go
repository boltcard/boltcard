package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
	"crypto/sha256"

	lnrpc "github.com/lightningnetwork/lnd/lnrpc"
	routerrpc "github.com/lightningnetwork/lnd/lnrpc/routerrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"gopkg.in/macaroon.v2"
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
		log.Printf("unable to connect to %s: %w", fullHostname, err)
		panic(err)
	}

	return connection
}

func add_invoice(amount_sat int64, metadata string) (payment_request string, return_err error) {

	ln_port, err := strconv.Atoi(os.Getenv("LN_PORT"))
	if err != nil {
		return "", err
	}

	dh := sha256.Sum256([]byte(metadata))

	connection := getGrpcConn(
		os.Getenv("LN_HOST"),
		ln_port,
		os.Getenv("LN_TLS_FILE"),
		os.Getenv("LN_MACAROON_FILE"))

	l_client := lnrpc.NewLightningClient(connection)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := l_client.AddInvoice(ctx, &lnrpc.Invoice {
		Value:			amount_sat,
		DescriptionHash:	dh[:],
	})

	if err != nil {
		return "", err
	}

	return result.PaymentRequest, nil
}

func pay_invoice(invoice string) (payment_status string, failure_reason string, return_err error) {

	payment_status = ""
	failure_reason = ""
	return_err = nil

	// SendPaymentV2

	// get node parameters from environment variables

	ln_port, err := strconv.Atoi(os.Getenv("LN_PORT"))
	if err != nil {
		return_err = err
		return
	}

	connection := getGrpcConn(
		os.Getenv("LN_HOST"),
		ln_port,
		os.Getenv("LN_TLS_FILE"),
		os.Getenv("LN_MACAROON_FILE"))

	r_client := routerrpc.NewRouterClient(connection)

	fee_limit_sat_str := os.Getenv("FEE_LIMIT_SAT")
	fee_limit_sat, err := strconv.ParseInt(fee_limit_sat_str, 10, 64)
	if err != nil {
		return_err = err
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := r_client.SendPaymentV2(ctx, &routerrpc.SendPaymentRequest{
		PaymentRequest:    invoice,
		NoInflightUpdates: true,
		TimeoutSeconds:    30,
		FeeLimitSat:       fee_limit_sat})

	if err != nil {
		return_err = err
		return
	}

	for {
		update, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			return_err = err
			return
		}

		payment_status = lnrpc.Payment_PaymentStatus_name[int32(update.Status)]
		failure_reason = lnrpc.PaymentFailureReason_name[int32(update.FailureReason)]
	}

	connection.Close()

	return
}
