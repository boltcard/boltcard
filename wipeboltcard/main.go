package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	qrcode "github.com/skip2/go-qrcode"
	"strconv"
)

type card_wipe_info struct {
	id  int
	k0  string
	k1  string
	k2  string
	k3  string
	k4  string
	uid string
}

func main() {

	card_name_ptr := flag.String("name", "", "select the card to be wiped by name")

	flag.Parse()

	if *card_name_ptr == "" {
		flag.PrintDefaults()
		return
	}

	// check if card_name exists

	card_count, err := db_get_card_name_count(*card_name_ptr)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	if card_count == 0 {
		fmt.Println("the card name does not exist in the database")
		return
	}

	// set the card as wiped and disabled, get the keys

	card_wipe_info_values, err := db_wipe_card(*card_name_ptr)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	// show a QR code on the console

	qr := `{` +
		`"action": "wipe",` +
		`"id": ` + strconv.Itoa(card_wipe_info_values.id) + `,` +
		`"k0": "` + card_wipe_info_values.k0 + `",` +
		`"k1": "` + card_wipe_info_values.k1 + `",` +
		`"k2": "` + card_wipe_info_values.k2 + `",` +
		`"k3": "` + card_wipe_info_values.k3 + `",` +
		`"k4": "` + card_wipe_info_values.k4 + `",` +
		`"uid": "` + card_wipe_info_values.uid + `",` +
		`"version": 1` +
		`}`

	fmt.Println()
	q, err := qrcode.New(qr, qrcode.Medium)
	fmt.Println(q.ToSmallString(false))
}
