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

func db_delete_expired() error {

	db, err := db_open()
	if err != nil {
		return err
	}
	defer db.Close()

	// delete expired one time code records

	sqlStatement := `DELETE FROM cards WHERE one_time_code_expiry < NOW() AND one_time_code_used = 'N';`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		return err
	}

	return nil
}

func db_insert_card(one_time_code string, lock_key string, aes_cmac string) error {

	db, err := db_open()
	if err != nil {
		return err
	}
	defer db.Close()

	// insert a new record into cards

	sqlStatement := `INSERT INTO cards` +
		` (one_time_code, lock_key, aes_cmac, uid, last_counter_value,` +
		` lnurlw_request_timeout_sec, tx_limit_sats, day_limit_sats, one_time_code_used)` +
		` VALUES ($1, $2, $3, '', 0, 60, 1000, 10000, 'N');`
	res, err := db.Exec(sqlStatement, one_time_code, lock_key, aes_cmac)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("not one card record inserted")
	}

	return nil
}
