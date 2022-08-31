package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	qrcode "github.com/skip2/go-qrcode"
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

	help_flag_ptr := flag.Bool("help", false, "show the command line options")
	tx_max_ptr := flag.Int("tx_max", 0, "set the maximum satoshis per transaction")
	day_max_ptr := flag.Int("day_max", 0, "set the maximum satoshis per day (24 hours)")
	enable_flag_ptr := flag.Bool("enable", false, "enable the card for payments")
	card_name_ptr := flag.String("name", "", "set a name for the card")

	flag.Parse()

	// handle -help

	if *help_flag_ptr {
		flag.PrintDefaults()
		return
	}

	fmt.Println()
	fmt.Println("use './createboltcard -help' to show command line options")

	// create the keys

	one_time_code := random_hex()
	k0_auth_key := random_hex()
	k2_cmac_key := random_hex()
	k3 := random_hex()
	k4 := random_hex()

	// create the new card record

	err := db_insert_card(one_time_code, k0_auth_key, k2_cmac_key, k3, k4,
		*tx_max_ptr, *day_max_ptr, *enable_flag_ptr, *card_name_ptr)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	// remove any expired records

	err = db_delete_expired()
	if err != nil {
		log.Warn(err.Error())
		return
	}

	// show a QR code on the console for the URI + one_time_code

	hostdomain := os.Getenv("HOST_DOMAIN")
	url := "https://" + hostdomain + "/new?a=" + one_time_code
	fmt.Println()
	fmt.Println(url)
	fmt.Println()
	q, err := qrcode.New(url, qrcode.Medium)
	fmt.Println(q.ToSmallString(false))
}
