# Steps for making a bolt card with the Android app

## Introduction

Here we describe how to create your own bolt cards with the Bolt Card service and the Bolt Card Android app.

## Resources
 
- some `NXP DNA 424 NTAG` cards
- an Android device with NFC
- a Bolt Card service
- [the Bolt Card app](https://github.com/boltcard/bolt-nfc-android-app)
- [the Bolt Card app usage document](https://github.com/boltcard/bolt-nfc-android-app#usage)

## Steps

### Install the app

- install the app from
  - source
  - apk
  - Google Play Store [Boltcard NFC Card Creator](https://play.google.com/store/apps/details?id=com.lightningnfcapp)

### Write the key values to the card
on the bolt card server
- ensure the environment variables for the database connection are set up (see `boltcard.service`)   
this can be achieved by writing these lines to the end of the `~/.bashrc` file  
```
echo "writing database_login to env vars"

export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=cardapp
export DB_PASSWORD=database_password
export DB_NAME=card_db

echo "writing host_domain to env vars"

export HOST_DOMAIN=card.yourdomain.com
```
- enter the `createboltcard` directory
- `$ go build`
- run the creation program
  - `./createboltcard -help` to see options
  - `./createboltcard -enable -tx_max=1000 -day_max=10000 -name=card_1` for example
- this will give you a one-time link in text and QR code form

on the app
- click `scan QR code`
- scan the QR code
- the app will prompt you to hold the card for programming
- the app will test the card and show you the results

### Make a payment
- monitor the bolt card service logs
- `$ journalctl -u boltcard.service -f`
- use a PoS setup to read the bolt card, e.g. [Breez wallet](https://breez.technology/)
