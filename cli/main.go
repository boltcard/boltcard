package main

import (
	"encoding/hex"
	"fmt"
	"github.com/boltcard/boltcard/crypto"
	"os"
)

// inspired by parse_request() in lnurlw_request.go

func check_cmac(uid []byte, ctr []byte, k2_cmac_key []byte, cmac []byte) (bool, error) {

	sv2 := make([]byte, 16)
	sv2[0] = 0x3c
	sv2[1] = 0xc3
	sv2[2] = 0x00
	sv2[3] = 0x01
	sv2[4] = 0x00
	sv2[5] = 0x80
	sv2[6] = uid[0]
	sv2[7] = uid[1]
	sv2[8] = uid[2]
	sv2[9] = uid[3]
	sv2[10] = uid[4]
	sv2[11] = uid[5]
	sv2[12] = uid[6]
	sv2[13] = ctr[2]
	sv2[14] = ctr[1]
	sv2[15] = ctr[0]

	cmac_verified, err := crypto.Aes_cmac(k2_cmac_key, sv2, cmac)

	if err != nil {
		return false, err
	}

	return cmac_verified, nil
}

func main() {

	fmt.Println("-- bolt card crypto test vectors --")
	fmt.Println()

	args := os.Args[1:]

	if len(args) != 4 {
		fmt.Println("error: should have arguments for: p c aes_decrypt_key aes_cmac_key")
		os.Exit(1)
	}

	// get from args
	p_hex := args[0]
	c_hex := args[1]
	aes_decrypt_key_hex := args[2]
	aes_cmac_key_hex := args[3]

	fmt.Println("p = ", p_hex)
	fmt.Println("c = ", c_hex)
	fmt.Println("aes_decrypt_key = ", aes_decrypt_key_hex)
	fmt.Println("aes_cmac_key = ", aes_cmac_key_hex)
	fmt.Println()

	p, err := hex.DecodeString(p_hex)

	if err != nil {
		fmt.Println("ERROR: p not valid hex", err)
		os.Exit(1)
	}

	c, err := hex.DecodeString(c_hex)

	if err != nil {
		fmt.Println("ERROR: c not valid hex", err)
		os.Exit(1)
	}

	if len(p) != 16 {
		fmt.Println("ERROR: p length not valid")
		os.Exit(1)
	}

	if len(c) != 8 {
		fmt.Println("ERROR: c length not valid")
		os.Exit(1)
	}

	// decrypt p with aes_decrypt_key

	aes_decrypt_key, err := hex.DecodeString(aes_decrypt_key_hex)

	if err != nil {
		fmt.Println("ERROR: DecodeString() returned an error", err)
		os.Exit(1)
	}

	dec_p, err := crypto.Aes_decrypt(aes_decrypt_key, p)

	if err != nil {
		fmt.Println("ERROR: Aes_decrypt() returned an error", err)
		os.Exit(1)
	}

	if dec_p[0] != 0xC7 {
		fmt.Println("ERROR: decrypted data does not start with 0xC7 so is invalid")
		os.Exit(1)
	}

	uid := dec_p[1:8]

	ctr := make([]byte, 3)
	ctr[0] = dec_p[10]
	ctr[1] = dec_p[9]
	ctr[2] = dec_p[8]

	// set up uid & ctr for card record if needed

	uid_str := hex.EncodeToString(uid)
	ctr_str := hex.EncodeToString(ctr)

	fmt.Println("decrypted card data : uid", uid_str, ", ctr", ctr_str)

	// check cmac

	aes_cmac_key, err := hex.DecodeString(aes_cmac_key_hex)

	if err != nil {
		fmt.Println("ERROR: aes_cmac_key is not valid hex", err)
		os.Exit(1)
	}

	cmac_valid, err := check_cmac(uid, ctr, aes_cmac_key, c)

	if err != nil {
		fmt.Println("ERROR: check_cmac() returned an error", err)
		os.Exit(1)
	}

	if cmac_valid == false {
		fmt.Println("ERROR: cmac incorrect")
		os.Exit(1)
	}

	fmt.Println("cmac validates ok")
	os.Exit(0)
}
