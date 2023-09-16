# test vectors

some test vectors to help with developing code to AES decode and validate lnurlw:// requests

these have been created by using an actual card and with [a small command line utility](https://github.com/boltcard/boltcard/blob/main/cli/main.go)

```
-- bolt card crypto test vectors --

p =  4E2E289D945A66BB13377A728884E867
c =  E19CCB1FED8892CE
aes_decrypt_key =  0c3b25d92b38ae443229dd59ad34b85d
aes_cmac_key =  b45775776cb224c75bcde7ca3704e933

decrypted card data : uid 04996c6a926980 , ctr 000003
sv2 =  [60 195 0 1 0 128 4 153 108 106 146 105 128 3 0 0]
ks =  [242 92 75 92 230 171 63 244 5 242 135 175 172 78 77 26]
cm =  [118 225 233 156 238 203 64 31 163 237 110 136 112 146 124 206]
ct =  [225 156 203 31 237 136 146 206]
cmac validates ok



-- bolt card crypto test vectors --

p =  00F48C4F8E386DED06BCDC78FA92E2FE
c =  66B4826EA4C155B4
aes_decrypt_key =  0c3b25d92b38ae443229dd59ad34b85d
aes_cmac_key =  b45775776cb224c75bcde7ca3704e933

decrypted card data : uid 04996c6a926980 , ctr 000005
sv2 =  [60 195 0 1 0 128 4 153 108 106 146 105 128 5 0 0]
ks =  [73 70 39 105 116 24 126 152 96 101 139 189 130 16 200 190]
cm =  [94 102 243 180 93 130 2 110 198 164 241 193 67 85 112 180]
ct =  [102 180 130 110 164 193 85 180]
cmac validates ok



-- bolt card crypto test vectors --

p =  0DBF3C59B59B0638D60B5842A997D4D1
c =  CC61660C020B4D96
aes_decrypt_key =  0c3b25d92b38ae443229dd59ad34b85d
aes_cmac_key =  b45775776cb224c75bcde7ca3704e933

decrypted card data : uid 04996c6a926980 , ctr 000007
sv2 =  [60 195 0 1 0 128 4 153 108 106 146 105 128 7 0 0]
ks =  [97 189 177 81 15 79 217 5 102 95 162 58 192 199 38 97]
cm =  [40 204 202 97 87 102 6 12 101 2 250 11 199 77 73 150]
ct =  [204 97 102 12 2 11 77 150]
cmac validates ok

```
