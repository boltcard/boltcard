package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	qrcode "github.com/skip2/go-qrcode"
	"os"
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
	one_time_code := random_hex()
	lock_key := random_hex()
	aes_cmac := random_hex()

	// create the new card record

	err := db_insert_card(one_time_code, lock_key, aes_cmac)
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
	fmt.Println(url)
	q, err := qrcode.New(url, qrcode.Medium)
	fmt.Println(q.ToSmallString(false))
}
