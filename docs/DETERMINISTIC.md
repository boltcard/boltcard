## Abstract

The NXP NTAG424DNA allows applications to configure five application keys, named K0, K1, K2, K3, and K4. In the Bolt card configuration:

* K0 is the only key permitted to change the application keys.
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
* `K0 = IssuerKey`.
* `K1 = PRF(K0, '2d003f77' || batchId)` with `batchId` being 4 bytes identifying the batch of card. (Can be set to `00000000` if uneeded)
* `K2 = PRF(K1, '2d003f78' || UID)`
* `K3 = PRF(K1, '2d003f79' || UID)`
* `K4 = PRF(K1, '2d003f7a' || UID)`

Under this proposed solution:
* With a card and the `Issuer Key`, the `Issuing Service` can recover all five application keys for that card.
* With a card and the `Encryption Key`, the `LNUrl Withdraw Service` can recover all application keys except for the `Issuer Key` (`K0`).
* The `Issuing Service` can reset any Bolt Card using only the `Issuer Key`.
* The `LNUrl Withdraw Service` might still need to brute-force encryption keys if there are multiple batches of Bolt Cards and no information in the lnurlw specifies to which batch a card belongs. However, this would require brute-forcing only one encryption key per batch, rather than one per card.

Note: Since both the `Encryption Key` (`K1`) and the `Issuer Key` (`K0`) are shared across a batch of cards, it's advisable to confirm that the card originates from NXP using `Read_Sig` before altering the keys.

## Test vectors

Input:
```
UID: 04a39493cc8680
Batch: 01000000
K0: 00000000000000000000000000000001
```

Expected:

```
K1: aa104a0bef8f751add9f06c5f000837a
K2: c98b6607222caffcac227f4f6241bd68
K3: d6e5ce82ec27f9d8c5d91d7c0c3a9f80
K4: d9352ff7ed7b43a13980a8c78aa4383a
```