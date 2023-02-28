package db

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

type Card struct {
	Card_id                    int
	Card_guid                  string
	K0_auth_key                string
	K1_decrypt_key             string
	K2_cmac_key                string
	K3                         string
	K4                         string
	Db_uid                     string
	Last_counter_value         uint32
	Lnurlw_request_timeout_sec int
	Lnurlw_enable              string
	Tx_limit_sats              int
	Day_limit_sats             int
	Lnurlp_enable              string
	Email_address              string
	Email_enable               string
	Uid_privacy                string
	One_time_code              string
	Card_name                  string
	Allow_negative_balance     string
}

type Payment struct {
	Card_payment_id int
	Card_id         int
	Lnurlw_k1       string
	Paid_flag       string
}

type Transaction struct {
	Card_id         int
	Tx_id           int
	Tx_type         string
	Tx_amount_msats int
	Tx_time         string
}

type Card_wipe_info struct {
	Id  int
	K0  string
	K1  string
	K2  string
	K3  string
	K4  string
	Uid string
}

func open() (*sql.DB, error) {

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

func Get_setting(setting_name string) string {

	setting_value := ""

	db, err := open()
	if err != nil {
		return ""
	}
	defer db.Close()

	sqlStatement := `select value from settings where name=$1;`

	row := db.QueryRow(sqlStatement, setting_name)
	err = row.Scan(&setting_value)
	if err != nil {
		return ""
	}

	return setting_value
}

func Get_new_card(one_time_code string) (*Card, error) {

	c := Card{}

	db, err := open()
	if err != nil {
		return &c, err
	}
	defer db.Close()

	sqlStatement := `SELECT k0_auth_key, k2_cmac_key, k3, k4, card_name, uid_privacy` +
		` FROM cards WHERE one_time_code=$1 AND` +
		` one_time_code_expiry > NOW() AND one_time_code_used = 'N' AND wiped = 'N';`
	row := db.QueryRow(sqlStatement, one_time_code)
	err = row.Scan(
		&c.K0_auth_key,
		&c.K2_cmac_key,
		&c.K3,
		&c.K4,
		&c.Card_name,
		&c.Uid_privacy)
	if err != nil {
		return &c, err
	}

	sqlStatement = `UPDATE cards SET one_time_code_used = 'Y' WHERE one_time_code = $1;`
	_, err = db.Exec(sqlStatement, one_time_code)
	if err != nil {
		return &c, err
	}

	return &c, nil
}

func Get_card_count_for_uid(uid string) (int, error) {

	card_count := 0

	db, err := open()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	sqlStatement := `select count(card_id) from cards where uid=$1 AND wiped='N';`

	row := db.QueryRow(sqlStatement, uid)
	err = row.Scan(&card_count)
	if err != nil {
		return 0, err
	}

	return card_count, nil
}

func Get_card_count_for_name_lnurlp(name string) (int, error) {

	card_count := 0

	db, err := open()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	sqlStatement := `select count(card_id) from cards where card_name=$1 and lnurlp_enable='Y';`

	row := db.QueryRow(sqlStatement, name)
	err = row.Scan(&card_count)
	if err != nil {
		return 0, err
	}

	return card_count, nil
}

// gets the last record
func Get_card_id_for_name(name string) (int, error) {

	card_id := 0

	db, err := open()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	sqlStatement := `select card_id from cards where card_name=$1 order by card_id desc limit 1;`

	row := db.QueryRow(sqlStatement, name)
	err = row.Scan(&card_id)
	if err != nil {
		return 0, err
	}

	return card_id, nil
}

func Get_card_id_for_card_payment_id(card_payment_id int) (int, error) {
	card_id := 0

	db, err := open()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	sqlStatement := `SELECT card_id FROM card_payments WHERE card_payment_id=$1;`

	row := db.QueryRow(sqlStatement, card_payment_id)
	err = row.Scan(&card_id)
	if err != nil {
		return 0, err
	}

	return card_id, nil
}

func Get_card_id_for_r_hash(r_hash string) (int, error) {
	card_id := 0

	db, err := open()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	sqlStatement := `SELECT card_id FROM card_receipts WHERE r_hash_hex=$1;`

	row := db.QueryRow(sqlStatement, r_hash)
	err = row.Scan(&card_id)
	if err != nil {
		return 0, err
	}

	return card_id, nil
}

func Get_cards_blank_uid() ([]Card, error) {

	// open the database

	db, err := open()

	if err != nil {
		return nil, err
	}

	defer db.Close()

	// query the database

	sqlStatement := `select card_id, k2_cmac_key from cards where uid='' and last_counter_value=0;`

	rows, err := db.Query(sqlStatement)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// prepare the results

	var cards []Card

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var c Card
		err := rows.Scan(&c.Card_id, &c.K2_cmac_key)

		if err != nil {
			return cards, err
		}
		cards = append(cards, c)
	}

	err = rows.Err()

	if err != nil {
		return cards, err
	}

	return cards, nil
}

func Update_card_uid_ctr(card_id int, uid string, ctr uint32) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStatement := `UPDATE cards SET uid = $2, last_counter_value = $3 WHERE card_id = $1;`
	res, err := db.Exec(sqlStatement, card_id, uid, ctr)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return nil
	}

	return nil
}

func Get_card_from_uid(card_uid string) (*Card, error) {

	c := Card{}

	db, err := open()
	if err != nil {
		return &c, err
	}
	defer db.Close()

	sqlStatement := `SELECT card_id, k2_cmac_key, uid,` +
		` last_counter_value, lnurlw_request_timeout_sec,` +
		` lnurlw_enable, tx_limit_sats, day_limit_sats` +
		` FROM cards WHERE uid=$1 AND wiped='N';`
	row := db.QueryRow(sqlStatement, card_uid)
	err = row.Scan(
		&c.Card_id,
		&c.K2_cmac_key,
		&c.Db_uid,
		&c.Last_counter_value,
		&c.Lnurlw_request_timeout_sec,
		&c.Lnurlw_enable,
		&c.Tx_limit_sats,
		&c.Day_limit_sats)
	if err != nil {
		return &c, err
	}

	return &c, nil
}

func Get_card_from_card_id(card_id int) (*Card, error) {

	c := Card{}

	db, err := open()
	if err != nil {
		return &c, err
	}
	defer db.Close()

	sqlStatement := `SELECT card_id, k2_cmac_key, uid, ` +
		`last_counter_value, lnurlw_request_timeout_sec, ` +
		`lnurlw_enable, tx_limit_sats, day_limit_sats, ` +
		`email_enable, email_address, card_name, ` +
		`allow_negative_balance FROM cards WHERE card_id=$1;`
	row := db.QueryRow(sqlStatement, card_id)
	err = row.Scan(
		&c.Card_id,
		&c.K2_cmac_key,
		&c.Db_uid,
		&c.Last_counter_value,
		&c.Lnurlw_request_timeout_sec,
		&c.Lnurlw_enable,
		&c.Tx_limit_sats,
		&c.Day_limit_sats,
		&c.Email_enable,
		&c.Email_address,
		&c.Card_name,
		&c.Allow_negative_balance)
	if err != nil {
		return &c, err
	}

	return &c, nil
}

// non wiped cards only
func Get_card_from_card_name(card_name string) (*Card, error) {

	c := Card{}

	db, err := open()
	if err != nil {
		return &c, err
	}
	defer db.Close()

	sqlStatement := `SELECT card_id, k2_cmac_key, uid,` +
		` last_counter_value, lnurlw_request_timeout_sec,` +
		` lnurlw_enable, tx_limit_sats, day_limit_sats` +
		` FROM cards WHERE card_name=$1 AND wiped = 'N';`
	row := db.QueryRow(sqlStatement, card_name)
	err = row.Scan(
		&c.Card_id,
		&c.K2_cmac_key,
		&c.Db_uid,
		&c.Last_counter_value,
		&c.Lnurlw_request_timeout_sec,
		&c.Lnurlw_enable,
		&c.Tx_limit_sats,
		&c.Day_limit_sats)
	if err != nil {
		return &c, err
	}

	return &c, nil
}

func Check_lnurlw_timeout(card_payment_id int) (bool, error) {

	db, err := open()
	if err != nil {
		return true, err
	}
	defer db.Close()

	lnurlw_timeout := true

	sqlStatement := `SELECT NOW() > cp.lnurlw_request_time + c.lnurlw_request_timeout_sec * INTERVAL '1 SECOND'` +
		` FROM  card_payments AS cp INNER JOIN cards AS c ON c.card_id = cp.card_id` +
		` WHERE cp.card_payment_id=$1;`
	row := db.QueryRow(sqlStatement, card_payment_id)
	err = row.Scan(&lnurlw_timeout)
	if err != nil {
		return true, err
	}

	return lnurlw_timeout, nil
}

func Check_and_update_counter(card_id int, new_counter_value uint32) (bool, error) {

	db, err := open()
	if err != nil {
		return false, err
	}
	defer db.Close()

	sqlStatement := `UPDATE cards SET last_counter_value = $2 WHERE card_id = $1` +
		` AND last_counter_value < $2;`
	res, err := db.Exec(sqlStatement, card_id, new_counter_value)
	if err != nil {
		return false, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if count != 1 {
		return false, nil
	}

	return true, nil
}

func Insert_payment(card_id int, lnurlw_k1 string) error {

	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	// insert a new record into card_payments with card_id & lnurlw_k1 set

	sqlStatement := `INSERT INTO card_payments` +
		` (card_id, lnurlw_k1, paid_flag, lnurlw_request_time, payment_status_time)` +
		` VALUES ($1, $2, 'N', NOW(), NOW());`
	res, err := db.Exec(sqlStatement, card_id, lnurlw_k1)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("not one card_payments record inserted")
	}

	return nil
}

func Insert_receipt(
	card_id int,
	ln_invoice string,
	r_hash_hex string,
	amount_msat int64) error {

	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	// insert a new record into card_receipts

	sqlStatement := `INSERT INTO card_receipts` +
		` (card_id, ln_invoice, r_hash_hex, amount_msats, receipt_status_time)` +
		` VALUES ($1, $2, $3, $4, NOW());`
	res, err := db.Exec(sqlStatement, card_id, ln_invoice, r_hash_hex, amount_msat)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("not one card_receipts record inserted")
	}

	return nil
}

func Update_receipt_state(r_hash_hex string, invoice_state string) error {
	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStatement := `UPDATE card_receipts ` +
		`SET receipt_status = $2, receipt_status_time = NOW() ` +
		`WHERE r_hash_hex = $1;`
	res, err := db.Exec(sqlStatement, r_hash_hex, invoice_state)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("not one card_receipts record updated")
	}

	return nil
}

func Get_payment_k1(lnurlw_k1 string) (*Payment, error) {
	p := Payment{}

	db, err := open()
	if err != nil {
		return &p, err
	}
	defer db.Close()

	sqlStatement := `SELECT card_payment_id, card_id, paid_flag` +
		` FROM card_payments WHERE lnurlw_k1=$1;`
	row := db.QueryRow(sqlStatement, lnurlw_k1)
	err = row.Scan(
		&p.Card_payment_id,
		&p.Card_id,
		&p.Paid_flag)
	if err != nil {
		return &p, err
	}

	return &p, nil
}

func Update_payment_invoice(card_payment_id int, ln_invoice string, amount_msats int64) error {

	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStatement := `UPDATE card_payments SET ln_invoice = $2, amount_msats = $3 WHERE card_payment_id = $1;`
	res, err := db.Exec(sqlStatement, card_payment_id, ln_invoice, amount_msats)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("not one card_payments record updated")
	}

	return nil
}

func Update_payment_paid(card_payment_id int) error {

	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	sqlStatement := `UPDATE card_payments SET paid_flag = 'Y', payment_time = NOW() WHERE card_payment_id = $1;`
	res, err := db.Exec(sqlStatement, card_payment_id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("not one card_payment record updated")
	}

	return nil
}

func Update_payment_status(card_payment_id int, payment_status string, failure_reason string) error {

	db, err := open()

	if err != nil {
		return err
	}

	defer db.Close()

	sqlStatement := `UPDATE card_payments SET payment_status = $2, failure_reason = $3, ` +
		`payment_status_time = NOW() WHERE card_payment_id = $1;`

	res, err := db.Exec(sqlStatement, card_payment_id, payment_status, failure_reason)

	if err != nil {
		return err
	}

	count, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if count != 1 {
		return errors.New("not one card_payment record updated")
	}

	return nil
}

func Get_card_totals(card_id int) (int, error) {

	db, err := open()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	day_total_msats := 0

	sqlStatement := `SELECT COALESCE(SUM(amount_msats),0) FROM card_payments ` +
		`WHERE card_id=$1 AND paid_flag='Y' ` +
		`AND payment_time > NOW() - INTERVAL '1 DAY';`
	row := db.QueryRow(sqlStatement, card_id)
	err = row.Scan(&day_total_msats)
	if err != nil {
		return 0, err
	}

	day_total_sats := day_total_msats / 1000

	return day_total_sats, nil
}

func Get_card_txs(card_id int, max_txs int) ([]Transaction, error) {
	// open the database

	db, err := open()

	if err != nil {
		return nil, err
	}

	defer db.Close()

	// query the database

	sqlStatement := `SELECT card_id, ` +
		`card_payments.card_payment_id AS tx_id, 'payment' AS tx_type, ` +
		`amount_msats as tx_amount_msats, ` +
		`TO_CHAR(payment_status_time, 'DD/MM/YYYY HH:MI:SS') AS tx_time ` +
		`FROM card_payments WHERE card_id = $1 AND payment_status != 'FAILED' ` +
		`AND payment_status != '' ` +
		`AND amount_msats != 0 UNION SELECT card_id, card_receipts.card_receipt_id AS tx_id, ` +
		`'receipt' AS tx_type, amount_msats as tx_amount_msats, ` +
		`TO_CHAR(receipt_status_time, 'DD/MM/YYYY HH:MI:SS') AS tx_time ` +
		`FROM card_receipts WHERE card_id = $1 ` +
		`AND receipt_status = 'SETTLED' ORDER BY tx_time DESC LIMIT $2`

	rows, err := db.Query(sqlStatement, card_id, max_txs)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// prepare the results

	var transactions []Transaction

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var t Transaction
		err := rows.Scan(&t.Card_id, &t.Tx_id, &t.Tx_type, &t.Tx_amount_msats, &t.Tx_time)

		if err != nil {
			return transactions, err
		}
		transactions = append(transactions, t)
	}

	err = rows.Err()

	if err != nil {
		return transactions, err
	}

	return transactions, nil
}

func Get_card_total_sats(card_id int) (int, error) {

	db, err := open()
	if err != nil {
		return 0, err
	}

	card_total_msats := 0

	sqlStatement := `SELECT COALESCE(SUM(tx_amount_msats),0) FROM (SELECT card_id, ` +
		`card_payments.card_payment_id AS tx_id, 'payment' AS tx_type, ` +
		`-amount_msats as tx_amount_msats, payment_status_time AS tx_time ` +
		`FROM card_payments WHERE card_id = $1 AND payment_status != 'FAILED' ` +
		`AND payment_status != '' ` +
		`AND amount_msats != 0 UNION SELECT card_id, card_receipts.card_receipt_id AS tx_id, ` +
		`'receipt' AS tx_type, amount_msats as tx_amount_msats, ` +
		`receipt_status_time AS tx_time FROM card_receipts WHERE card_id = $1 ` +
		`AND receipt_status = 'SETTLED' ORDER BY tx_time) AS transactions;`

	row := db.QueryRow(sqlStatement, card_id)
	err = row.Scan(&card_total_msats)
	if err != nil {
		return 0, err
	}

	card_total_sats := card_total_msats / 1000

	return card_total_sats, nil
}

func Get_card_name_count(card_name string) (card_count int, err error) {

	card_count = 0

	db, err := open()
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

func Insert_card(one_time_code string, k0_auth_key string, k2_cmac_key string, k3 string, k4 string,
	tx_limit_sats int, day_limit_sats int, lnurlw_enable bool, card_name string, uid_privacy bool,
	allow_neg_bal_ptr bool) error {

	lnurlw_enable_yn := "N"
	if lnurlw_enable {
		lnurlw_enable_yn = "Y"
	}

	uid_privacy_yn := "N"
	if uid_privacy {
		uid_privacy_yn = "Y"
	}

	allow_neg_bal_yn := "N"
	if allow_neg_bal_ptr {
		allow_neg_bal_yn = "Y"
	}

	db, err := open()
	if err != nil {
		return err
	}
	defer db.Close()

	// ensure any cards with the same card_name are wiped

	sqlStatement := `UPDATE cards SET` +
		` lnurlw_enable = 'N', lnurlp_enable = 'N', email_enable = 'N', wiped = 'Y'` +
		` WHERE card_name = $1;`
	res, err := db.Exec(sqlStatement, card_name)
	if err != nil {
		return err
	}

	// insert a new record into cards

	sqlStatement = `INSERT INTO cards` +
		` (one_time_code, k0_auth_key, k2_cmac_key, k3, k4, uid, last_counter_value,` +
		` lnurlw_request_timeout_sec, tx_limit_sats, day_limit_sats, lnurlw_enable,` +
		` one_time_code_used, card_name, uid_privacy, allow_negative_balance)` +
		` VALUES ($1, $2, $3, $4, $5, '', 0, 60, $6, $7, $8, 'N', $9, $10, $11);`
	res, err = db.Exec(sqlStatement, one_time_code, k0_auth_key, k2_cmac_key, k3, k4,
		tx_limit_sats, day_limit_sats, lnurlw_enable_yn, card_name, uid_privacy_yn,
		allow_neg_bal_yn)
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

func Wipe_card(card_name string) (*Card_wipe_info, error) {

	card_wipe_info := Card_wipe_info{}

	db, err := open()
	if err != nil {
		return &card_wipe_info, err
	}
	defer db.Close()

	// set card as wiped and disabled

	sqlStatement := `UPDATE cards SET` +
		` lnurlw_enable = 'N', lnurlp_enable = 'N', email_enable = 'N', wiped = 'Y'` +
		` WHERE card_name = $1;`
	_, err = db.Exec(sqlStatement, card_name)
	if err != nil {
		return &card_wipe_info, err
	}

	// get card keys for the last card wiped

	sqlStatement = `SELECT card_id, uid, k0_auth_key, k2_cmac_key, k3, k4` +
		` FROM cards WHERE card_name = $1 ORDER BY card_id DESC LIMIT 1;`
	row := db.QueryRow(sqlStatement, card_name)
	err = row.Scan(
		&card_wipe_info.Id,
		&card_wipe_info.Uid,
		&card_wipe_info.K0,
		&card_wipe_info.K2,
		&card_wipe_info.K3,
		&card_wipe_info.K4)
	if err != nil {
		return &card_wipe_info, err
	}

	card_wipe_info.K1 = Get_setting("AES_DECRYPT_KEY")

	return &card_wipe_info, nil
}

func Update_card(card_name string, lnurlw_enable bool, tx_limit_sats int, day_limit_sats int) error {

	lnurlw_enable_yn := "N"
	if lnurlw_enable {
		lnurlw_enable_yn = "Y"
	}

	db, err := open()

	if err != nil {
		return err
	}

	defer db.Close()

	sqlStatement := `UPDATE cards SET lnurlw_enable = $2, tx_limit_sats = $3, day_limit_sats = $4 ` +
		`WHERE card_name = $1 AND wiped = 'N';`

	res, err := db.Exec(sqlStatement, card_name, lnurlw_enable_yn, tx_limit_sats, day_limit_sats)

	if err != nil {
		return err
	}

	count, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if count != 1 {
		return errors.New("not one card record updated")
	}

	return nil
}
