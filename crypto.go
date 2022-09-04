package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"

	"github.com/aead/cmac"
)

func create_k1() (string, error) {

	// 16 bytes = 128 bits
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	str := hex.EncodeToString(b)

	return str, nil
}

// decrypt p with aes_dec
func crypto_aes_decrypt(key_sdm_file_read []byte, ba_p []byte) ([]byte, error) {

	dec_p := make([]byte, 16)
	iv := make([]byte, 16)
	c1, err := aes.NewCipher(key_sdm_file_read)
	if err != nil {
		return dec_p, err
	}
	mode := cipher.NewCBCDecrypter(c1, iv)
	mode.CryptBlocks(dec_p, ba_p)

	return dec_p, nil
}

func crypto_aes_cmac(key_sdm_file_read_mac []byte, sv2 []byte, ba_c []byte) (bool, error) {

	c2, err := aes.NewCipher(key_sdm_file_read_mac)
	if err != nil {
		return false, err
	}
	ks, err := cmac.Sum(sv2, c2, 16)
	if err != nil {
		return false, err
	}
	c3, err := aes.NewCipher(ks)
	if err != nil {
		return false, err
	}
	cm, err := cmac.Sum([]byte{}, c3, 16)
	if err != nil {
		return false, err
	}
	ct := make([]byte, 8)
	ct[0] = cm[1]
	ct[1] = cm[3]
	ct[2] = cm[5]
	ct[3] = cm[7]
	ct[4] = cm[9]
	ct[5] = cm[11]
	ct[6] = cm[13]
	ct[7] = cm[15]

	res_cmac := bytes.Compare(ct, ba_c)
	if res_cmac != 0 {
		return false, nil
	}

	return true, nil
}
