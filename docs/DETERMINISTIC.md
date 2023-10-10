## Abstract

The NXP NTAG424DNA allows applications to configure five application keys, named K0, K1, K2, K3, and K4. In the Bolt card configuration:

* K0 is the `App Master Key`, it is the only key permitted to change the application keys.
* K1 serves as the `encryption key` for the PICC Data, represented by the `p=` parameter.
* K2 is the `authentication key` for the PICC Data, represented by the `c=` parameter.
* K3 and K4 are not used but should be configured as recommended in the application notes.

A simplistic approach to issuing Bolt cards would involve randomly generating the five different keys and storing them in a database.

When a validation request is made, the verifier would attempt to decrypt the `p=` parameter using all existing encryption keys until finding a match. Once decrypted, the `p=` parameter would reveal the card's uid, which can then be used to retrieve the remaining keys.

The primary drawback of this method is its lack of scalability. If many cards have been issued, identifying the correct encryption key could become computationally expensive.

In this document, we propose a solution to this issue.

## Key generation

First, it's important to understand that a Bolt Card issuer consists of two distinct services:
* `Issuing Service`: This agent sets up the cards for lightning payments, which involves specifying a particular `LNUrl Withdraw Service` and generating the application keys.
* `LNUrl Withdraw Service`: This service authenticates the card and completes the payment.

Assuming the `Issuing Service` generates a random key named (the `Issuer Key`) and has a batch of Bolt Cards to configure, it will set the following parameters:
* `K0 = PRF(IssuerKey, '2d003f76' || UID)`
* `K1 = PRF(IssuerKey, '2d003f77' || batchId)` with `batchId` being 4 bytes identifying the batch of card. (Can be set to `00000000` if uneeded)
* `K2 = PRF(K1, '2d003f78' || UID)`
* `K3 = PRF(K1, '2d003f79' || UID)`
* `K4 = PRF(K1, '2d003f7a' || UID)`

The Pseudo Random Function `PRF(key, message)` applied during the key generation is the CMAC algorithm described in NIST Special Publication 800-38B.

Under this proposed solution:
* With a card and the `Issuer Key`, the `Issuing Service` can recover all five application keys for that card.
* With a card and the `Encryption Key`, the `LNUrl Withdraw Service` can recover all application keys except for the `Issuer Key` (`K0`).
* The `Issuing Service` can reset any Bolt Card using only the `Issuer Key`.
* The `LNUrl Withdraw Service` might still need to brute-force encryption keys if there are multiple batches of Bolt Cards and no information in the lnurlw specifies to which batch a card belongs. However, this would require brute-forcing only one encryption key per batch, rather than one per card.

## How the to implement a Reset feature

If an `Issuing Service` offers a factory reset feature for a user's bolt card, here is the recommended procedure:

1. Read the NDEF lnurlw URL.
2. Brute-force the encryption and authentication of the card using all existing `batchId` values to find `K1`, `K2`, and the `UID`.
3. Use the `UID` from the `PICCData`, along with `K1` and the `IssuerKey`, to recover `K0`, `K3`, and `K4`.
5. Execute `AuthenticateEV2First` with `K0`
6. Erase the NDEF data file using `WriteData` or `ISOUpdateBinary`
7. Restore the NDEF file settings to default values with `ChangeFileSettings`.
8. Use `ChangeKey` with the recovered application keys to reset `K4` through `K0` to `00000000000000000000000000000000`.

## How to implement a verification

If a `LNUrl Withdraw Service` needs to verify a payment request, follow these steps:

1. Read the NDEF lnurlw URL.
2. Brute-force the encryption and authentication of the card using all existing `batchId` values to find `K1`, `K2` and `UID`.
3. Confirm that the last-seen counter for `ID=PRF(K1, '2d003f7b' || UID)[0..7]` is lower than what is stored in the `PICCData`.
4. Update the last-seen counter.

Note that `LNUrl Withdraw Service` can't derive `App Master Key` (`K0`), and thus is unable to change the keys of the bolt card.

The specific method for calculating `ID` is not crucial; the recommendation is to avoid using `UID` directly. This approach offers both privacy and security benefits.

Firstly, since the `UID` is used to derive keys, it should not be stored outside the NTag.

Secondly, this allows a user to re-flash the same NTag with a different `batchId` or through a different `Issuing Service`, letting the user to obtain a different `ID` for the same NTag.

Third, this prevent tracking of the NTag across different `Issuing Service`.

## Security consideration

Since `K0` and `K1` are shared among multiple Bolt Cards, the security of this scheme is based on the following assumptions:

* `K0` and `K1` cannot be extracted from a legitimate NTag424.
* Bolt Card setup occurs in a trusted environment.

While NXP gives assurance keys can't be extracted, a non genuine NTag424 could potentially expose these keys.

Furthermore, because Bolt Card setup uses the well-known initial application keys `00000000000000000000000000000000`, communication between the PCD and the PICC could be intercepted. If the Bolt Card setup doesn't occurs in a trusted environment, `K0` and `K1` could be exposed during the calls to `ChangeKey`.

Note that verifying the signature returned by `Read_Sig` can only prove NXP issued a card with a specific `UID`. It cannot prove that the current communication channel is established with an authentic NTag424. This is because the signature returned by `Read_Sig` covers only the UID and can therefore be replayed by a non-genuine NTag424.

## Test vectors

Input:
```
UID: 04a39493cc8680
Batch: 01000000
Issuer Key: 00000000000000000000000000000001
```

Expected:

```
K0: 75da58a68fbb1bef64708e87c7be9ad3
K1: aa104a0bef8f751add9f06c5f000837a
K2: c98b6607222caffcac227f4f6241bd68
K3: d6e5ce82ec27f9d8c5d91d7c0c3a9f80
K4: d9352ff7ed7b43a13980a8c78aa4383a
ID: a98da306ba6d90
```