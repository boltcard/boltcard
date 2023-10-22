## Abstract

The NXP NTAG424DNA allows applications to configure five application keys, named `K0`, `K1`, `K2`, `K3`, and `K4`. In the Bolt card configuration:

* `K0` is the `App Master Key`, it is the only key permitted to change the application keys.
* `K1` serves as the `encryption key` for the `PICCData`, represented by the `p=` parameter.
* `K2` is the `authentication key` used for calculating the SUN MAC of the `PICCData`, represented by the `c=` parameter.
* `K3` and `K4` are not used but should be configured as recommended in the [NTag424 application notes](https://www.nxp.com/docs/en/application-note/AN12196.pdf).

A simplistic approach to issuing Bolt cards would involve randomly generating the five different keys and storing them in a database.

When a validation request is made, the verifier would attempt to decrypt the `p=` parameter using all existing encryption keys until finding a match. Once decrypted, the `p=` parameter would reveal the card's uid, which can then be used to retrieve the remaining keys.

The primary drawback of this method is its lack of scalability. If many cards have been issued, identifying the correct encryption key could become computationally expensive.

In this document, we propose a solution to this issue.

## Key generation

Assuming the `LNUrl Withdraw Service` generates a random key named (the `IssuerKey`) and has a `batch` of Bolt Cards to configure, it will set the following parameters:

* `K0 = PRF(IssuerKey, '2d003f76' || batchId || UID)`
* `K1 = PRF(IssuerKey, '2d003f77' || batchId)`
* `K2 = PRF(IssuerKey, '2d003f78' || batchId || UID)`
* `K3 = PRF(IssuerKey, '2d003f79' || batchId || UID)`
* `K4 = PRF(IssuerKey, '2d003f7a' || batchId || UID)`

`batchId`: 4 bytes identifying the batch of card. (Can be set to `00000000` if uneeded)

The Pseudo Random Function `PRF(key, message)` applied during the key generation is the CMAC algorithm described in NIST Special Publication 800-38B.

## How the to implement a Reset feature

If a `LNUrl Withdraw Service` offers a factory reset feature for a user's bolt card, here is the recommended procedure:

1. Read the NDEF lnurlw URL, extract `p=` and `c=`.
2. For each existing `batchId`:
    1. Derive `K1`, decrypts `p=` to get the `PICCData`.
    2. If `PICCData[0] != 0xc7`, go to the next `batchId`.
    3. Take `UID=PICCData[1..8]`, derive `K2`
    4. Calculate the SUN MAC with `K2`, if different from `c=`, go to next `batchId`
3. From the `UID`, the `IssuerKey` and the `batchId` with correct SUN MAC, recover `K0`, `K3`, and `K4`.
5. Execute `AuthenticateEV2First` with `K0`
6. Erase the NDEF data file using `WriteData` or `ISOUpdateBinary`
7. Restore the NDEF file settings to default values with `ChangeFileSettings`.
8. Use `ChangeKey` with the recovered application keys to reset `K4` through `K0` to `00000000000000000000000000000000`.

Rational: Attempting to call `AuthenticateEV2First` without validating the `p=` and `c=` parameters could render the NTag inoperable after a few attempts.

## How to implement a verification

If a `LNUrl Withdraw Service` needs to verify a payment request, follow these steps:

1. Read the NDEF lnurlw URL, extract `p=` and `c=`.
2. For each existing `batchId`:
    1. Derive `K1`, decrypts `p=` to get the `PICCData`.
    2. If `PICCData[0] != 0xc7`, go to the next `batchId`.
    3. Take `UID=PICCData[1..8]`, derive `K2`
    4. Calculate the SUN MAC with `K2`, if different from `c=`, go to next `batchId`
3. If no correct SUN MAC has been found, returns an error.
3. Confirm that the last-seen counter for `ID=PRF(IssuerKey, '2d003f7b' || batchId || UID)[0..7]` is lower than what is stored in `counter=PICCData[8..11]`.
4. Update the last-seen counter.

The specific method for calculating `ID` is not crucial; the recommendation is to avoid using `UID` directly. This approach offers both privacy and security benefits.

Mainly, since the `UID` is used to derive keys, it is better to not store it outside the NTag.

## Security consideration

Since `K1` is shared among multiple Bolt Cards, the security of this scheme is based on the following assumptions:

* `K1` cannot be extracted from a legitimate NTag424.
* Bolt Card setup occurs in a trusted environment.

While NXP gives assurance keys can't be extracted, a non genuine NTag424 could potentially expose these keys.

Furthermore, because blank NTag424 uses the well-known initial application keys `00000000000000000000000000000000`, communication between the PCD and the PICC could be intercepted. If the Bolt Card setup doesn't occurs in a trusted environment, `K1` could be exposed during the calls to `ChangeKey`.

However, if `K1` is compromised, the attacker still cannot produce a valid checksum and can only recover the `UID` for tracking purposes.

Note that verifying the signature returned by `Read_Sig` can only prove NXP issued a card with a specific `UID`. It cannot prove that the current communication channel is established with an authentic NTag424. This is because the signature returned by `Read_Sig` covers only the `UID` and can therefore be replayed by a non-genuine NTag424.

## Test vectors

Input:
```
UID: 04a39493cc8680
Batch: 01000000
Issuer Key: 00000000000000000000000000000001
```

Expected:

```
K0: 60ef62b99ed8dc351ef7382b7d9e60f0
K1: aa104a0bef8f751add9f06c5f000837a
K2: 0365b383bafe15365289939d9631d6b2
K3: fb753c7436da79395278f13d4aa0a406
K4: 8e069871bd7c2f0c9d2ce8ffba54e4c7
ID: d702d970ac2b3f
```
