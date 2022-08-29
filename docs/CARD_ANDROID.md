# Steps for making a bolt card with the Android app

## Introduction

Here we describe how to create your own bolt cards with the Bolt Card Android app and the Bolt Card service.

## Resources
 
- some `NXP DNA 424 NTAG` cards
- an Android device with NFC
- a Bolt Card service
- [the Bolt Card app](https://github.com/boltcard/bolt-nfc-android-app)
- [the Bolt Card app usage document](https://github.com/boltcard/bolt-nfc-android-app#usage)

## Steps

### Install the app

- install the app from source or apk

### Write the URI template to the card
on the app
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

### Write the key values to the card
on the bolt card server
- ensure the environment variables for the database connection are set up (see `boltcard.service`)    
- enter the `createboltcard` directory
- `$ go build`
- `./createboltcard` to create a card
- `./createboltcard -help` to see options
- `./createboltcard -enable -tx_max=1000 -day_max=10000 -name=card_1` for example
- this will give you a one-time link in text and QR code form
- if the boltcard service is running in **docker**, use ```docker exec boltcard_main createboltcard/createboltcard``` instead

on the app
- select `Key Management`
- click `scan QR code from console`
- scan the QR code
- bring the card to the device for programming the keys

### Update the card record on the server
on the bolt card db server
- `$ psql card_db`
- `card_db=# select card_id, one_time_code from cards order by card_id desc limit 1;`
- check that this is the correct record (one_time_code matches from before)
- `card_db=# update cards set uid = 'UID value from before without the 0x prefix' where card_id=card_id from before;`
- `card_db=# update cards set enable_flag = 'Y' where card_id=card_id from before;`

### Make a payment
- monitor the bolt card service logs
- `$ journalctl -u boltcard.service -f`
- use a PoS setup to read the bolt card, e.g. [Breez wallet](https://breez.technology/)
