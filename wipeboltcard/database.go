package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

func db_open() (*sql.DB, error) {

	// get connection string from environment variables

	conn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	db, err := sql.Open("postgres", conn)
	if err != nil {
		return db, err
	}

	return db, nil
}

func db_get_card_name_count(card_name string) (card_count int, err error) {

        card_count = 0

        db, err := db_open()
        if err != nil {
                return 0, err
        }
        defer db.Close()

        sqlStatement := `SELECT COUNT(card_id) FROM cards WHERE card_name = $1;`

        row := db.QueryRow(sqlStatement, card_name)
        err = row.Scan(&card_count)
        if err != nil {
                return 0, err
        }

        return card_count, nil
}

func db_wipe_card(card_name string) (*card_wipe_info, error) {

	card_wipe_info := card_wipe_info{}

	db, err := db_open()
	if err != nil {
		return &card_wipe_info, err
	}
	defer db.Close()

	// set card as wiped and disabled

	sqlStatement := `UPDATE cards SET` +
		` lnurlw_enable = 'N', lnurlp_enable = 'N', email_enable = 'N', wiped = 'Y'` +
		` WHERE card_name = $1;`
	res, err := db.Exec(sqlStatement, card_name)
	if err != nil {
		return &card_wipe_info, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return &card_wipe_info, err
	}
	if count != 1 {
		return &card_wipe_info, errors.New("not one card record updated")
	}

	// get card keys

	sqlStatement = `SELECT card_id, uid, k0_auth_key, k2_cmac_key, k3, k4` +
		` FROM cards WHERE card_name = $1;`
	row := db.QueryRow(sqlStatement, card_name)
	err = row.Scan(
		&card_wipe_info.id,
		&card_wipe_info.uid,
		&card_wipe_info.k0,
		&card_wipe_info.k2,
		&card_wipe_info.k3,
		&card_wipe_info.k4)
	if err != nil {
		return &card_wipe_info, err
	}

	card_wipe_info.k1 = db_get_setting("AES_DECRYPT_KEY")

	return &card_wipe_info, nil
}
