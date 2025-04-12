# NeuroData Cryptographic Formats Specification

:: type: Specification
:: version: 0.1.0
:: status: draft
:: dependsOn: docs/metadata.md, docs/references.md, docs/neurodata_and_composite_file_spec.md
:: howToUpdate: Refine attributes for each type, specify encoding formats (PEM, Base64, Hex), detail tool requirements and expected behaviors.

## 1. Purpose

This document defines a set of NeuroData formats for representing common cryptographic objects: identities/key pairs (`.ndid`), digital certificates (`.ndcert`), digital signatures (`.ndsig`), and cryptographic hashes (`.ndhash`). These formats aim to store cryptographic information and metadata in a human-readable structure suitable for use within the NeuroScript ecosystem, primarily facilitating verification, identification, and integrity checks performed by dedicated tools.

## 2. Security Considerations

* **Private Keys:** These formats are **NOT** intended for storing sensitive private key material directly in plain text. Private keys should be managed through secure external mechanisms (e.g., hardware security modules, encrypted storage, environment variables). References (like `PRIVATE_KEY_REF`) may point to identifiers for such keys, but the key data itself should generally not be present. Storing public keys, certificates, signatures, and hashes is the primary goal.
* **Tooling:** The usefulness of these formats is entirely dependent on the existence and correct implementation of secure cryptographic tools within the NeuroScript environment (see Section 5).

## 3. Common Elements

All cryptographic formats below generally follow these conventions:
* **Structure:** Tagged Line Structure (similar to `.ndform`), where each object starts with metadata, followed by attribute lines defining its properties.
* **Metadata:** Each format uses standard `:: key: value` metadata, including `:: type: <TypeName>`, `:: version:`, and potentially `:: id:`.
* **Data Encoding:** Large opaque data like public keys, certificates, or signatures are typically stored within fenced blocks (e.g., ```pem ... ```, ```base64 ... ```, ```hex ... ```) immediately following their corresponding attribute tag (e.g., `PUBLIC_KEY`, `CERTIFICATE`, `SIGNATURE`, `HASH_VALUE`). The language tag of the fence indicates the encoding. Simple hashes or fingerprints might be represented as quoted strings directly on the attribute line.
* **References:** Uses the standard `[ref:<location>#<block_id>]` or `[ref:<location>]` syntax [cite: generated previously in `docs/references.md`] to link related objects (e.g., a signature referencing the content it signs, a certificate referencing its public key identity).

## 4. Format Specifications

### 4.1 Identity / Key Pair (`.ndid`)

* **Purpose:** Represents a cryptographic identity, primarily its public key.
* `:: type: Identity`
* **Attributes:**
    * `LABEL "<name>"`: Human-readable name.
    * `TYPE "<algorithm>"`: E.g., "RSA", "Ed25519", "ECDSA".
    * `SIZE <bits>`: Key size (e.g., 2048, 256).
    * `PUBLIC_KEY` (Followed by ` ```pem ... ``` ` block containing the public key).
    * `PRIVATE_KEY_REF "<identifier_or_uri>"`: (Optional, points to secure storage, NOT the key itself).
    * `CREATED_AT <timestamp>`: (Optional).
    * `DESCRIPTION "<text>"`: (Optional).
* **Example:**
    ```ndid
    :: type: Identity
    :: version: 0.1.0
    :: id: service-key-001

    LABEL "Main Service Signing Key"
    TYPE "Ed25519"
    PUBLIC_KEY ```pem
    -----BEGIN PUBLIC KEY-----
    MCowBQYDK2VwAyEA[...]EXAMPLE[...]oW1lA=
    -----END PUBLIC KEY-----
    ```
    CREATED_AT 2024-03-01T12:00:00Z
    ```

### 4.2 Certificate (`.ndcert`)

* **Purpose:** Represents a digital certificate (typically X.509).
* `:: type: Certificate`
* **Attributes:**
    * `SUBJECT "<distinguished_name>"`
    * `ISSUER "<distinguished_name>"`
    * `SERIAL_NUMBER "<hex_or_decimal_string>"`
    * `VERSION <number>` (e.g., 3 for X.509v3)
    * `VALID_FROM <timestamp>`
    * `VALID_UNTIL <timestamp>`
    * `SIGNATURE_ALGORITHM "<name_or_oid>"`
    * `PUBLIC_KEY_INFO` (Optional, followed by ` ```text ... ``` ` block with details like type/size, or use `PUBLIC_KEY_REF`)
    * `PUBLIC_KEY_REF "[ref:...]"` (Optional link to an `.ndid` object).
    * `FINGERPRINT_SHA256 "<hex_string>"` (Optional, common identifier).
    * `CERTIFICATE` (Followed by ` ```pem ... ``` ` block containing the certificate).
* **Example:**
    ```ndcert
    :: type: Certificate
    :: version: 0.1.0
    :: id: web-server-cert-2025

    SUBJECT "CN=[www.example.com](https://www.example.com), O=Example Corp, C=US"
    ISSUER "CN=Example Intermediate CA G1, O=Example Corp, C=US"
    SERIAL_NUMBER "1A:2B:3C:4D:..."
    VERSION 3
    VALID_FROM 2024-01-01T00:00:00Z
    VALID_UNTIL 2025-01-01T23:59:59Z
    SIGNATURE_ALGORITHM "SHA256withRSA"
    FINGERPRINT_SHA256 "a1b2c3d4e5f6..."
    PUBLIC_KEY_REF "[ref:keys/webserver.ndid#key-01]"
    CERTIFICATE ```pem
    -----BEGIN CERTIFICATE-----
    MIIDqDCCApCgAwIBAgIJA[...]EXAMPLE[...]K4A/DQ==
    -----END CERTIFICATE-----
    ```
    ```

### 4.3 Signature (`.ndsig`)

* **Purpose:** Represents a digital signature over specified content.
* `:: type: Signature`
* **Attributes:**
    * `CONTENT_REF "[ref:<location>[#<block_id>]]"`: Reference to the signed content (file or block).
    * `SIGNER_REF "[ref:<location>#<key_id>]"`: Reference to the signer's `.ndid` public key.
    * `ALGORITHM "<sig_alg_name>"`: E.g., "SHA256withRSA", "EdDSA", "ECDSAwithSHA256".
    * `TIMESTAMP <iso8601_timestamp>`: (Optional) Time of signing.
    * `SIGNATURE` (Followed by ` ```base64 ... ``` ` or ` ```hex ... ``` ` block containing the signature value).
* **Example:**
    ```ndsig
    :: type: Signature
    :: version: 0.1.0
    :: id: manifest-sig-v2

    CONTENT_REF "[ref:manifest.json]"
    SIGNER_REF "[ref:identities/release_key.ndid#prod-signer]"
    ALGORITHM "EdDSA"
    TIMESTAMP 2024-04-12T14:30:00Z
    SIGNATURE ```base64
    K7L[...]EXAMPLE[...]9wA=
    ```
    ```

### 4.4 Hash (`.ndhash`)

* **Purpose:** Represents a cryptographic hash of specified content.
* `:: type: Hash`
* **Attributes:**
    * `CONTENT_REF "[ref:<location>[#<block_id>]]"`: Reference to the hashed content.
    * `ALGORITHM "<hash_alg_name>"`: E.g., "SHA-256", "SHA-512", "BLAKE2b-256".
    * `HASH_VALUE "<hex_string>"`: The hash value, typically hex encoded. (Could allow Base64 block too).
* **Example:**
    ```ndhash
    :: type: Hash
    :: version: 0.1.0

    CONTENT_REF "[ref:firmware.bin]"
    ALGORITHM "SHA-256"
    HASH_VALUE "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    ```

## 5. Tooling Requirements

Using these formats effectively requires a suite of cryptographic tools within the NeuroScript environment. These tools currently **do not exist** [cite: uploaded:neuroscript/pkg/core/tools_register.go] and would need to be implemented, potentially wrapping standard Go cryptographic libraries.

Potential required tools:
* `TOOL.GenerateKeyPair(type, size)` -> `ndid_content`
* `TOOL.CalculateHash(content_ref_or_string, algorithm)` -> `hash_value_string`
* `TOOL.CreateSignature(content_ref_or_string, private_key_ref, algorithm)` -> `signature_value_base64`
* `TOOL.VerifySignature(signature_ref, public_key_ref_or_ndid, algorithm)` -> `bool` (Verifies against the content referenced *within* the `.ndsig` file/block).
* `TOOL.ParseCertificate(cert_ref_or_content)` -> `map` (Extracts fields from a certificate).
* `TOOL.ValidateCertificate(cert_ref, trusted_roots_ref)` -> `bool` (Checks validity period, signature, potentially chain).
* `TOOL.GetPublicKeyFromCert(cert_ref_or_content)` -> `ndid_content` (Extracts public key from cert).
* `TOOL.EncodeData(data, format)` -> `string` (e.g., format="base64", "hex", "pem")
* `TOOL.DecodeData(encoded_string, format)` -> `data`

Implementing these tools securely requires careful handling of keys and cryptographic primitives.
