# test vectors

some test vectors to help with developing code to AES decode and validate lnurlw:// requests

these have been created by using an actual card and with [a small command line utility](https://github.com/boltcard/boltcard/blob/main/cli/main.go)

```
p =  4E2E289D945A66BB13377A728884E867
c =  E19CCB1FED8892CE
aes_decrypt_key =  0c3b25d92b38ae443229dd59ad34b85d
aes_cmac_key =  b45775776cb224c75bcde7ca3704e933

decrypted card data : uid 04996c6a926980 , ctr 000003
cmac validates ok



-- bolt card crypto test vectors --

p =  00F48C4F8E386DED06BCDC78FA92E2FE
c =  66B4826EA4C155B4
aes_decrypt_key =  0c3b25d92b38ae443229dd59ad34b85d
aes_cmac_key =  b45775776cb224c75bcde7ca3704e933

decrypted card data : uid 04996c6a926980 , ctr 000005
cmac validates ok



-- bolt card crypto test vectors --

p =  0DBF3C59B59B0638D60B5842A997D4D1
c =  CC61660C020B4D96
aes_decrypt_key =  0c3b25d92b38ae443229dd59ad34b85d
aes_cmac_key =  b45775776cb224c75bcde7ca3704e933

decrypted card data : uid 04996c6a926980 , ctr 000007
cmac validates ok

```
