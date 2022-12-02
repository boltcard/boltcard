package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	qrcode "github.com/skip2/go-qrcode"
	"os"
	"strings"
)

func random_hex() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Warn(err.Error())
		return ""
	}

	return hex.EncodeToString(b)
}

func main() {

	tx_max_ptr := flag.Int("tx_max", 0, "set the maximum satoshis per transaction")
	day_max_ptr := flag.Int("day_max", 0, "set the maximum satoshis per 24 hour day")
	enable_flag_ptr := flag.Bool("enable", false, "enable the card for payments")
	card_name_ptr := flag.String("name", "", "set a name for the card (must be set)")
	uid_privacy_ptr := flag.Bool("uid_privacy", false, "select enhanced privacy for the card (cannot undo)")
	allow_neg_bal_ptr := flag.Bool("allow_neg_bal", false, "allow the card to have a negative balance")

	flag.Parse()

	if *card_name_ptr == "" {
		flag.PrintDefaults()
		return
	}

	// check if card_name already exists

	card_count, err := db_get_card_name_count(*card_name_ptr)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	if card_count > 0 {
		fmt.Println("the card name already exists in the database")
		return
	}

	// create the keys

	one_time_code := random_hex()
	k0_auth_key := random_hex()
	k2_cmac_key := random_hex()
	k3 := random_hex()
	k4 := random_hex()

	// create the new card record

	err = db_insert_card(one_time_code, k0_auth_key, k2_cmac_key, k3, k4,
		*tx_max_ptr, *day_max_ptr, *enable_flag_ptr, *card_name_ptr,
		*uid_privacy_ptr, *allow_neg_bal_ptr)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	// show a QR code on the console for the URI + one_time_code

	hostdomain := os.Getenv("HOST_DOMAIN")
	url := ""
	if strings.HasSuffix(hostdomain, ".onion") {
		url = "http://" + hostdomain + "/new?a=" + one_time_code
	} else {
		url = "https://" + hostdomain + "/new?a=" + one_time_code
	}

	fmt.Println()
	fmt.Println(url)
	fmt.Println()
	q, err := qrcode.New(url, qrcode.Medium)
	fmt.Println(q.ToSmallString(false))
}
