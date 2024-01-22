## Abstract

Boltcard NFC Programmer App is a native app on iOS and Android to flash or reset NTag424 into a Boltcard.

1. The `Boltcard Service` generates the keys, and format them into a QR Code
2. The user opens the Boltcard NFC Programmer, go to `Create Bolt Card`, scans the QR code
3. The user then taps the card

The QR code contains all the keys necessary for the app to create the Boltcard.

Here are the shortcomings we aim to address in this specification:

1. If the QR code is on the mobile device itself, it isn't possible to scan it
2. It isn't possible to generate a pair of keys specific for the NTag424 being setup (the [deterministic key generation](./DETERMINISTIC.md) needs the UID before generating the keys)

## Boltcard deeplinks

The solution is for the `Boltcard Service` to generate deep links with the following format: `Boltcard://[program|reset]?url=[keys-request-url]`.

When clicked, `Boltcard NFC Programmer` would open and either allow the user to program their NTag424 or reset it after asking for the NTags keys to the `keys-request-url`.

The `Boltcard NFC Programmer` should send an HTTP POST request with `Content-Type: application/json` in the following format:

```json
{
    "UID": "[UID]"
}
```

Or

```json
{
    "LNURLW": "lnurlw://..."
}
```

In `curl`:

```bash
curl -X POST "[keys-request-url]" -H "Content-Type: application/json" -d '{"UID": "[UID]"}'
```

* `UID` needs to be 7 bytes. (Program action)
* `LNURLW` needs to be read from the Boltcard's NDEF and can be sent in place of `UID`. It must contains the `p=` and `c=` arguments of the Boltcard. (Reset action)

The response will be similar to the format of the QR code:

```json
{
  "LNURLW": "lnurlw://...",
  "K0":"[Key0]",
  "K1":"[Key1]",
  "K2":"[Key2]",
  "K3":"[Key3]",
  "K4":"[Key4]"
}
```

## The Program action

If `program` is specified in the Boltcard link, the `Boltcard NFC Programmer` must:

1. Check if the lnurlw `NDEF` can be read.
    * If the record can be read, then the card isn't reset, an error should be displayed to the user to first reset the Boltcard with the previous `Boltcard Service`.
    * If the record can't be read, assume `K0` is `00000000000000000000000000000000` authenticate and call `GetUID` on the card again. (Since `GetUID` is called after authentication, the real `UID` will be returned even if `Random UID` has been activated)
2. Send a request to the `keys-request-url` using the UID as explained above to get the NTag424 app keys
3. Program the Boltcard

## The Reset action

If `reset` is specified in the Boltcard link, the `Boltcard NFC Programmer` must:
1. Check if the lnurlw `NDEF` can be read.
    * If the record can't be read, then the card is already reset, show an error message to the user.
    * If the record can be read, continue to step 2.
2. Send a request to the `keys-request-url` using the lnurlw as explained above to get the NTag424 app keys
3. Reset the Boltcard to factory state

## Handling setup/reset cycles for Boltcard Services

When a NTag424 is reset, its counter is reset too.
This means that if the user:

* Setup a Boltcard
* Make `5` payments
* Reset the Boltcard
* Setup the Boltcard on same `keys-request-url`

With a naive implementation, the server will expect the next counter to be above `5`, but the next payment will have a counter of `0`.

More precisely, the user will need to tap the card `5` times before being able to use the Boltcard for a payment successfully again.

To avoid this issue the `Boltcard Service`, if using [Deterministic key generation](./DETERMINISTIC.md), should ensure it updates the key version during a `program` action.

This can be done easily by the `Boltcard Service` by adding a parameter in the `keys-request-url` which specifies that the version need to be updated.

When the `Boltcard NFC Programmer` queries the URL with the UID of the card, the `Boltcard Service` will detect this parameter, and update the version.

## Test vectors

Here is an example of two links for respectively program the Boltcard and Reset it.

```html
<p>
    <a id="SetupBoltcard" href="boltcard://program?url=https%3A%2F%2Flocalhost%3A14142%2Fapi%2Fv1%2Fpull-payments%2FfUDXsnySxvb5LYZ1bSLiWzLjVuT%2Fboltcards%3FonExisting%3DUpdateVersion" target="_blank">
        Setup Boltcard
    </a>
    <span>&nbsp;|&nbsp;</span>
    <a id="ResetBoltcard" href="boltcard://reset?url=https%3A%2F%2Flocalhost%3A14142%2Fapi%2Fv1%2Fpull-payments%2FfUDXsnySxvb5LYZ1bSLiWzLjVuT%2Fboltcards%3FonExisting%3DKeepVersion" target="_blank">
        Reset Boltcard
    </a>
</p>
```