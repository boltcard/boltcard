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

## Keys generation

First, the `LNUrl Withdraw Service` generates a `IssuerKey` that it will use to generate the keys for every NTag424.

Then configure a Boltcard the following way:

* `CardKey = GetRandomBytes(16)`
* `K0 = PRF(CardKey, '2d003f76' || UID)`
* `K1 = PRF(IssuerKey, '2d003f77')`
* `K2 = PRF(CardKey, '2d003f78' || UID)`
* `K3 = PRF(CardKey, '2d003f79' || UID)`
* `K4 = PRF(CardKey, '2d003f7a' || UID)`

* `UID`: This is the 7-byte ID of the card. You can retrieve it from the NTag424 using the `GetCardUID` function after identification with K1, or by decrypting the `p=` parameter, also known as `PICCData`.

The Pseudo Random Function `PRF(key, message)` applied during the key generation is the CMAC algorithm described in NIST Special Publication 800-38B. [See implementation notes](#notes)

## How to setup a new boltcard

1. Generate a random `CardKey` of 16 bytes.
2. `ReadData` or `ISOReaDBinary` on the boltcard, to make sure the card is blank.
3. Execute `AuthenticateEV2First` with `00000000000000000000000000000000`
4. Fetch the `UID` with `GetCardUID`.
2. Calculate `K0`, `K1`, `K2`, `K3`, `K4`.
4. [Setup the boltcard](./CARD_MANUAL.md).

## How to implement a Reset feature

If a `LNUrl Withdraw Service` offers a factory reset feature for a user's bolt card, here is the recommended procedure:

1. Read the NDEF lnurlw URL, extract `p=` and `c=`.
2. Derive `Encryption Key (K1)`, decrypts `p=` to get the `PICCData`.
3. Check `PICCData[0] == 0xc7`. 
4. Calculate `ID=PRF(IssuerKey, '2d003f7b' || UID)` with the `UID` from the `PICCData`.
5. Fetch `CardKey` from database with `ID`.
6. Derive `K0`, `K2`, `K3`, `K4` with `CardKey` and the `UID`.
7. Verify that the SUN MAC in `c=` matches the one calculated using `Authentication Key (K2)`.
8. Execute `AuthenticateEV2First` with `K0`
9. Erase the NDEF data file using `WriteData` or `ISOUpdateBinary`
10. Restore the NDEF file settings to default values with `ChangeFileSettings`.
11. Use `ChangeKey` with the recovered application keys to reset `K4` through `K0` to `00000000000000000000000000000000`.

Rational: Attempting to call `AuthenticateEV2First` without validating the `p=` and `c=` parameters could render the NTag inoperable after a few attempts.

## How to implement a verification

If a `LNUrl Withdraw Service` needs to verify a payment request, follow these steps:

1. Read the NDEF lnurlw URL, extract `p=` and `c=`.
2. Derive `Encryption Key (K1)`, decrypts `p=` to get the `PICCData`.
3. Check `PICCData[0] == 0xc7`. 
4. Calculate `ID=PRF(IssuerKey, '2d003f7b' || UID)` with the `UID` from the `PICCData`.
5. Fetch `CardKey` from database with `ID`.
6. Derive `Authentication Key (K2)` with `CardKey` and the `UID`.
7. Verify that the SUN MAC in `c=` matches the one calculated using `Authentication Key (K2)`.
8. Confirm that the last-seen counter for `ID` is lower than what is stored in `counter=PICCData[8..11]`. (Little Endian)
9. Update the last-seen counter.

Rationale: The `ID` is calculated to prevent the exposure of the `UID` in the `LNUrl Withdraw Service` database. This approach provides both privacy and security. Specifically, because the `UID` is used to derive keys, it is preferable not to store it outside the NTag.

## Multiple IssuerKeys

A single `LNUrl Withdraw Service` can own multiple `IssuerKeys`. In such cases, it will need to attempt them all to decrypt `p=`, and pick the first one which satisfies `PICCData[0] == 0xc7` and verifies the `c=` checksum.

Using multiple `IssuerKeys`, can decrease the impact of a compromised `Encryption Key (K1)` at the cost of performance.

## Security consideration

### K1 security

Since `K1` is shared among multiple Bolt Cards, the security of this scheme is based on the following assumptions:

* `K1` cannot be extracted from a legitimate NTag424.
* Bolt Card setup occurs in a trusted environment.

While NXP gives assurance keys can't be extracted, a non genuine NTag424 could potentially expose these keys.

Furthermore, because blank NTag424 uses the well-known initial application keys `00000000000000000000000000000000`, communication between the PCD and the PICC could be intercepted. If the Bolt Card setup doesn't occurs in a trusted environment, `K1` could be exposed during the calls to `ChangeKey`.

However, if `K1` is compromised, the attacker still cannot produce a valid checksum and can only recover the `UID` for tracking purposes.

Note that verifying the signature returned by `Read_Sig` can only prove NXP issued a card with a specific `UID`. It cannot prove that the current communication channel is established with an authentic NTag424. This is because the signature returned by `Read_Sig` covers only the `UID` and can therefore be replayed by a non-genuine NTag424.

### Issuer database security

If the issuer's database is compromised, revealing both the IssuerKey and CardKeys, it would still be infeasible for an attacker to derive `K2` and thus to forge signatures for an arbitrary card.

This is because the database only stores `ID=PRF(IssuerKey, '2d003f7b' || UID)` and not the `UID` itself.

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

* [BTCPayServer.BoltCardTools](https://github.com/btcpayserver/BTCPayServer.BoltCardTools), a Boltcard/NTag424 library in C#.

## Test vectors

Input:
```
UID: 04a39493cc8680
Issuer Key: 00000000000000000000000000000001
Card Key: 00000000000000000000000000000002
```

Expected:
```
K0: 21940feffa2437910d8eb62b3b0a0648
K1: 55da174c9608993dc27bb3f30a4a7314
K2: 2934c4ab339979142dfd50ae0ca55dc2
K3: b696f18e5a79e5a0defb25c38109b8e3
K4: c9d493b9d3e62ce963586aafcd7c6cfe
ID: e07ce1279d980ecb892a81924b67bf18
```