## Abstract

The NXP NTAG424DNA allows applications to configure five application keys, named `K0`, `K1`, `K2`, `K3`, and `K4`. In the BoltCard configuration:

* `K0` is the `App Master Key`, it is the only key permitted to change the application keys.
* `K1` serves as the `encryption key` for the `PICCData`, represented by the `p=` parameter.
* `K2` is the `authentication key` used for calculating the SUN MAC of the `PICCData`, represented by the `c=` parameter.
* `K3` and `K4` are not used but should be configured as recommended in the [NTag424 application notes](https://www.nxp.com/docs/en/application-note/AN12196.pdf).

A simple approach to issuing BoltCards would involve randomly generating the five different keys and storing them in a database.

When a validation request is made, the verifier would attempt to decrypt the `p=` parameter using all existing encryption keys until finding a match. Once decrypted, the `p=` parameter would reveal the card's uid, which can then be used to retrieve the remaining keys.

The primary drawback of this method is its lack of scalability. If many cards have been issued, identifying the correct encryption key could become a computationally intensive task.

In this document, we propose a solution to this issue.

## Keys generation

First, the `LNUrl Withdraw Service` generates a `IssuerKey` that it will use to generate the keys for every NTag424.

Then, configure a BoltCard as follows:

* `CardKey = PRF(IssuerKey, '2d003f75' || UID || Version)`
* `K0 = PRF(CardKey, '2d003f76')`
* `K1 = PRF(IssuerKey, '2d003f77')`
* `K2 = PRF(CardKey, '2d003f78')`
* `K3 = PRF(CardKey, '2d003f79')`
* `K4 = PRF(CardKey, '2d003f7a')`
* `ID = PRF(IssuerKey, '2d003f7b' || UID)`

With the following parameters:
* `IssuerKey`: This 16-bytes key is used by an `LNUrl Withdraw Service` to setup all its BoltCards.
* `UID`: This is the 7-byte ID of the card. You can retrieve it from the NTag424 using the `GetCardUID` function after identification with K1, or by decrypting the `p=` parameter, also known as `PICCData`.
* `Version`: A 4-bytes little endian version number. This must be incremented every time the user re-programs (reset/setup) the same BoltCard on the same `LNUrl Withdraw Service`.

The Pseudo Random Function `PRF(key, message)` applied during the key generation is the CMAC algorithm described in NIST Special Publication 800-38B. [See implementation notes](#notes)

## How to setup a new BoltCard

1. Execute `ReadData` or `ISOReaDBinary` on the BoltCard to ensure the card is blank.
2. Execute `AuthenticateEV2First` with the application key `00000000000000000000000000000000`
3. Fetch the `UID` with `GetCardUID`.
4. Calculate `ID`
5. Fetch the `State` and `Version` of the BoltCard with the specified `ID` from the database.
6. Ensure either:
    * If no BoltCard is found, insert an entry in the database with `Version=0` and its state set to `Configured`.
    * If a BoltCard is found and its state is `Reset` then increment `Version` by `1`, and change its state to `Configured`.
7. Generate `CardKey` with `UID` and `Version`.
8. Calculate `K0`, `K1`, `K2`, `K3`, `K4`.
9. [Setup the BoltCard](./CARD_MANUAL.md).

## How to implement a Reset feature

If a `LNUrl Withdraw Service` offers a factory reset feature for a user's BoltCard, here is the recommended procedure:

1. Read the NDEF lnurlw URL, extract `p=` and `c=`.
2. Derive `Encryption Key (K1)`, decrypt `p=` to obtain the `PICCData`.
3. Check `PICCData[0] == 0xc7`. 
4. Calculate `ID` with the `UID` from the `PICCData`.
5. Fetch the BoltCard's `Version` with `ID` from the database.
6. Ensure the BoltCard's state is `Configured`.
7. Generate `CardKey` with `UID` and `Version`.
8. Derive `K0`, `K2`, `K3`, `K4` with `CardKey` and the `UID`.
9. Verify that the SUN MAC in `c=` matches the one calculated using `Authentication Key (K2)`.
10. Execute `AuthenticateEV2First` with `K0`
11. Erase the NDEF data file using `WriteData` or `ISOUpdateBinary`
12. Restore the NDEF file settings to default values with `ChangeFileSettings`.
13. Use `ChangeKey` with the recovered application keys to reset `K4` through `K0` to `00000000000000000000000000000000`.
14. Update the BoltCard's state to `Reset` in the database.

Rational: Attempting to call `AuthenticateEV2First` without validating the `p=` and `c=` parameters could render the NTag inoperable after a few attempts.

## How to implement a verification

If a `LNUrl Withdraw Service` needs to verify a payment request, follow these steps:

1. Read the NDEF lnurlw URL, extract `p=` and `c=`.
2. Derive `Encryption Key (K1)`, decrypts `p=` to get the `PICCData`.
3. Check `PICCData[0] == 0xc7`. 
4. Calculate `ID` with the `UID` from the `PICCData`.
5. Fetch the BoltCard's `Version` with `ID` from the database.
6. Ensure the BoltCard's state in the database is not `Reset`.
7. Generate `CardKey` with `UID` and `Version`.
8. Derive `Authentication Key (K2)` with `CardKey` and the `UID`.
9. Verify that the SUN MAC in `c=` matches the one calculated using `Authentication Key (K2)`.
10. Confirm that the last-seen counter for `ID` is lower than what is stored in `counter=PICCData[8..11]`. (Little Endian)
11. Update the last-seen counter.

Rationale: The `ID` is calculated to prevent the exposure of the `UID` in the `LNUrl Withdraw Service` database. This approach provides both privacy and security. Specifically, because the `UID` is used to derive keys, it is preferable not to store it outside the NTag.

## Multiple IssuerKeys

A single `LNUrl Withdraw Service` can own multiple `IssuerKeys`. In such cases, it will need to attempt them all to decrypt `p=`, and pick the first one which satisfies `PICCData[0] == 0xc7` and verifies the `c=` checksum.

Using multiple `IssuerKeys` can decrease the impact of a compromised `Encryption Key (K1)` at the cost of performance.

## Security consideration

### K1 security

Since `K1` is shared among multiple BoltCards, the security of this scheme is based on the following assumptions:

* `K1` cannot be extracted from a legitimate NTag424.
* BoltCard setup occurs in a trusted environment.

While NXP gives assurance keys can't be extracted, a non genuine NTag424 could potentially expose these keys.

Furthermore, because blank NTag424 uses the well-known initial application keys `00000000000000000000000000000000`, communication between the PCD and the PICC could be intercepted. If the BoltCard setup does not occur in a trusted environment, `K1` could be exposed during the calls to `ChangeKey`.

However, if `K1` is compromised, the attacker still cannot produce a valid checksum and can only recover the `UID` for tracking purposes.

Note that verifying the signature returned by `Read_Sig` can only prove NXP issued a card with a specific `UID`. It cannot prove that the current communication channel is established with an authentic NTag424. This is because the signature returned by `Read_Sig` covers only the `UID` and can therefore be replayed by a non-genuine NTag424.

### Issuer database security

If the issuer's database is compromised, revealing both the IssuerKey and CardKeys, it would still be infeasible for an attacker to derive `K2` and thus to forge signatures for an arbitrary card.

This is because the database only stores `ID` and not the `UID` itself.

## Implementation notes {#notes}

Here is a C# implementation of the CMAC algorithm described in NIST Special Publication 800-38B.

```csharp
public byte[] CMac(byte[] data)
{
    var key = _bytes;
    // SubKey generation
    // step 1, AES-128 with key K is applied to an all-zero input block.
    byte[] L = AesEncrypt(key, new byte[16], new byte[16]);

    // step 2, K1 is derived through the following operation:
    byte[]
        FirstSubkey =
            RotateLeft(L); //If the most significant bit of L is equal to 0, K1 is the left-shift of L by 1 bit.
    if ((L[0] & 0x80) == 0x80)
        FirstSubkey[15] ^=
            0x87; // Otherwise, K1 is the exclusive-OR of const_Rb and the left-shift of L by 1 bit.

    // step 3, K2 is derived through the following operation:
    byte[]
        SecondSubkey =
            RotateLeft(FirstSubkey); // If the most significant bit of K1 is equal to 0, K2 is the left-shift of K1 by 1 bit.
    if ((FirstSubkey[0] & 0x80) == 0x80)
        SecondSubkey[15] ^=
            0x87; // Otherwise, K2 is the exclusive-OR of const_Rb and the left-shift of K1 by 1 bit.

    // MAC computing
    if (((data.Length != 0) && (data.Length % 16 == 0)) == true)
    {
        // If the size of the input message block is equal to a positive multiple of the block size (namely, 128 bits),
        // the last block shall be exclusive-OR'ed with K1 before processing
        for (int j = 0; j < FirstSubkey.Length; j++)
            data[data.Length - 16 + j] ^= FirstSubkey[j];
    }
    else
    {
        // Otherwise, the last block shall be padded with 10^i
        byte[] padding = new byte[16 - data.Length % 16];
        padding[0] = 0x80;

        data = data.Concat(padding.AsEnumerable()).ToArray();

        // and exclusive-OR'ed with K2
        for (int j = 0; j < SecondSubkey.Length; j++)
            data[data.Length - 16 + j] ^= SecondSubkey[j];
    }

    // The result of the previous process will be the input of the last encryption.
    byte[] encResult = AesEncrypt(key, new byte[16], data);

    byte[] HashValue = new byte[16];
    Array.Copy(encResult, encResult.Length - HashValue.Length, HashValue, 0, HashValue.Length);

    return HashValue;
}
static byte[] RotateLeft(byte[] b)
{
    byte[] r = new byte[b.Length];
    byte carry = 0;

    for (int i = b.Length - 1; i >= 0; i--)
    {
        ushort u = (ushort)(b[i] << 1);
        r[i] = (byte)((u & 0xff) + carry);
        carry = (byte)((u & 0xff00) >> 8);
    }

    return r;
}
```

## Implementation

* [BTCPayServer.BoltCardTools](https://github.com/btcpayserver/BTCPayServer.BoltCardTools), a BoltCard/NTag424 library in C#.

## Test vectors

Input:
```
UID: 04a39493cc8680
Issuer Key: 00000000000000000000000000000001
Version: 1
```

Expected:
```
K0: a29119fcb48e737d1591d3489557e49b
K1: 55da174c9608993dc27bb3f30a4a7314
K2: f4b404be700ab285e333e32348fa3d3b
K3: 73610ba4afe45b55319691cb9489142f
K4: addd03e52964369be7f2967736b7bdb5
ID: e07ce1279d980ecb892a81924b67bf18
CardKey: ebff5a4e6da5ee14cbfe720ae06fbed9
```