# Steps for making a bolt card with the Android app

## Introduction

Here we describe how to create your own bolt card.

## Resources
 
- an Android device with NFC
- the BoltCard app
- [app usage document](https://github.com/boltcard/bolt-nfc-android-app#usage)

## Steps

### Install the app

[Bolt Card Android app](https://github.com/boltcard/bolt-nfc-android-app)  
- install the app from source or apk  

### Write to the card
- open the app
- select `Write NFC` 
- enter your domain and path in the text entry box given
```
card.yourdomain.com/ln
```
- bring the card to the device for programming the URI template  

- select `Read NFC`
- check that the URI looks correct
```
lnurlw://card.yourdomain.com/ln?c=...&p=...
```
- note the UID value

### Get the keys from the host
- on the bolt card server
- enter the `createboltcard` directory
- `$ go build`
- `./createboltcard`
- this will give you a one use link in text and QR code form  

- on the app
- select `Key Management`
- click `scan QR code from console`
- scan the QR code
- bring the card to the device for programming the keys

### Update the host card record
- on the bolt card server
- `$ psql card_db`
- `card_db=# select card_id, one_time_code from cards order by card_id desc limit 1;`
- check that this is the correct record
- `card_db=# update cards set uid = '*UID value from before without the 0x prefix*' where card_id=*card_id from before*;`
- `card_db=# update cards set enabled = 'Y' where card_id=*card_id from before*;`

### Make a payment
- monitor the bolt card service logs
- `$ journalctl -u boltcard.service -f`
- use a PoS setup to read the bolt card, e.g. [Breez wallet](https://breez.technology/)
- for more support options, see [Bolt card service installation](INSTALL.md)
