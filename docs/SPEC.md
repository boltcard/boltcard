# Bolt card specification

The bolt card system is built on the open standards listed below.

- [LUD-03: withdrawRequest base spec.](https://github.com/fiatjaf/lnurl-rfc/blob/luds/03.md)
- [LUD-17: Protocol schemes and raw (non bech32-encoded) URLs.](https://github.com/fiatjaf/lnurl-rfc/blob/luds/17.md)

## Bolt card and POS interaction

- the point-of-sale (POS) will read an NDEF message from the card, for example
```
lnurlw://card.yourdomain.com?p=A2EF40F6D46F1BB36E6EBF0114D4A464&c=F509EEA788E37E32
```
- the POS will call your bolt card service here
```
https://card.yourdomain.com?p=A2EF40F6D46F1BB36E6EBF0114D4A464&c=F509EEA788E37E32
```
- your bolt card service should verify the payment request and continue the LNURLw protocol

## Server side verification

- for the `p` value and the `SDM Meta Read Access Key` value, decrypt the UID and counter with AES
- for the `c` value and the `SDM File Read Access Key` value, check with AES-CMAC

- the authenticated UID and counter is used on the bolt card service to verify that the request is valid
- the bolt card service must only accept an increasing counter value
- additional validation rules can be added at the bolt card service, for example
  - card enable flag
  - card payment limit per transaction
  - card payment limit per day
  - allowed merchant list
  - verification of your location from your phone
- the bolt card service can then make payment from a connected lightning node
