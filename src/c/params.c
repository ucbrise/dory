#include <stdio.h>
#include <stdint.h>
#include <openssl/evp.h>
#include <openssl/sha.h>
#include <openssl/hmac.h>
#include <openssl/bn.h>
#include <string.h>
#include "common.h"
#include "params.h"
#include <math.h>

#define BYTE_TO_BINARY_PATTERN "%c%c%c%c%c%c%c%c"
#define BYTE_TO_BINARY(byte)  \
  (byte & 0x80 ? '1' : '0'), \
  (byte & 0x40 ? '1' : '0'), \
  (byte & 0x20 ? '1' : '0'), \
  (byte & 0x10 ? '1' : '0'), \
  (byte & 0x08 ? '1' : '0'), \
  (byte & 0x04 ? '1' : '0'), \
  (byte & 0x02 ? '1' : '0'), \
  (byte & 0x01 ? '1' : '0') 

/* Set parameters for entire system. */
void setSystemParams(int bloomFilterSz, int numDocs) {
    BLOOM_FILTER_SZ = bloomFilterSz;
    BLOOM_FILTER_BYTES = bloomFilterSz / 8;
    LOG_BLOOM_FILTER_SZ = ((int) log2((double)bloomFilterSz));
    MAX_DOCS = numDocs;
    MAX_DOCS_BYTES = numDocs / 8;
    NUM_DOCS = 0;
    NUM_DOCS_BYTES = 0;
    MALICIOUS_DPF_LEN = NUM_DOCS_BYTES + MAC_BYTES;
}

int min (int a, int b) {
  return (a < b) ? a : b;
}

/*
 * Use SHA-256 to hash the string in `bytes_in`
 * with the integer given in `counter`.
 */
int hashOnce (EVP_MD_CTX *ctx, uint8_t *bytes_out, 
    const uint8_t *bytes_in, int inlen, uint16_t counter) 
{
  int rv = ERROR;
  CHECK_C (EVP_DigestInit_ex (ctx, EVP_sha256 (), NULL));
  CHECK_C (EVP_DigestUpdate (ctx, &counter, sizeof counter));
  CHECK_C (EVP_DigestUpdate (ctx, bytes_in, inlen));
  CHECK_C (EVP_DigestFinal_ex (ctx, bytes_out, NULL));

cleanup:
  return rv; 
}

/*
 * Output a string of pseudorandom bytes by hashing a 
 * counter with the bytestring provided:
 *    Hash(0|bytes_in) | Hash(1|bytes_in) | ... 
 */
int hashToBytes (uint8_t *bytesOut, int outLen,
    const uint8_t *bytesIn, int inLen)
{
  int rv = ERROR;
  uint16_t counter = 0;
  uint8_t buf[SHA256_DIGEST_LENGTH];
  EVP_MD_CTX *ctx;

  ctx = EVP_MD_CTX_create();
  int bytesFilled = 0;
  do {
    const int toCopy = min (SHA256_DIGEST_LENGTH, outLen - bytesFilled);
    CHECK_C (hashOnce (ctx, buf, bytesIn, inLen, counter));
    memcpy (bytesOut + bytesFilled, buf, toCopy);
    
    counter++;
    bytesFilled += SHA256_DIGEST_LENGTH;
  } while (bytesFilled < outLen);

cleanup:
  if (ctx) EVP_MD_CTX_destroy(ctx);
  return rv; 
}

/* Evaluate entire PRF. */
int prf(EVP_CIPHER_CTX *ctx, uint8_t *bytesOut, int outLen,
        const uint8_t *bytesIn, int inLen, uint8_t startCounter) {
    int rv = ERROR;
    int bytesFilled = 0;
    uint8_t buf[16];
    uint8_t input[16];
    uint8_t counter = startCounter;

    memset(input, 0, 16);
    memcpy(input + 1, bytesIn, min(inLen, 15));

    do {
        memcpy(input, &counter, sizeof(uint8_t));
        int bytesCopied;
        int toCopy = min(16, outLen - bytesFilled);
        CHECK_C (EVP_EncryptUpdate(ctx, buf, &bytesCopied, input, 16));
        memcpy(bytesOut + bytesFilled, buf, toCopy);
        bytesFilled += 16;
        counter++;
    } while (bytesFilled < outLen);

cleanup:
    return rv;
}

/* Evaluate single block of PRF. */
int blockPrf(EVP_CIPHER_CTX *ctx, uint8_t *bytesOut,
        const uint8_t *bytesIn, int len) {
    int rv = ERROR;
    int bytesFilled = 0;
    int bytesCopied = 0;

    do {
        CHECK_C (EVP_EncryptUpdate(ctx, bytesOut, &bytesCopied, bytesIn, len)); 
        bytesFilled += bytesCopied;
    } while (bytesFilled < len);

cleanup:
    return rv;
}

/* Check if bit == 1. */
bool isBitOne(const uint8_t *buf, int bitIndex) {
    int byteIndex = bitIndex / 8;
    return buf[byteIndex] & (1 << (bitIndex % 8));
}

/* Set bit in buf to 1. */
void setBitOne(uint8_t *buf, int bitIndex) {
    int byteIndex = bitIndex / 8;
    buf[byteIndex] |= 1 << (bitIndex % 8);
}

/* Copy bit from dst to src. */
void copyBit(uint8_t *dst, int dstBitIndex, uint8_t *src, int srcBitIndex) {
    uint8_t srcBit = src[srcBitIndex / 8] & (1 << (srcBitIndex % 8));
    srcBit = srcBit >> (srcBitIndex % 8);
    dst[dstBitIndex / 8] |= srcBit << (dstBitIndex % 8);
}

/* XOR in into out. */
void xorIn(uint8_t *out, uint8_t *in, int len) {
    for (int i = 0; i < len; i++) {
        out[i] = out[i] ^ in[i];
    }
}

/* Extract a column of bits from a table. */
void getBitColumn(const uint8_t **table, int column, uint8_t *result) {
    int columnByte = column / 8;
    int columnIndex = column % 8;
    memset(result, 0, NUM_DOCS_BYTES);
    for (int i = 0; i < NUM_DOCS; i++) {
	    result[i/8] |= ((table[i][columnByte] & (1 << columnIndex)) >> columnIndex) << (i % 8);
    }
}

/* Set a bit column in the table. */
void setBitColumn(uint8_t **table, int column, uint8_t *in, int inBytes) {
    int columnByte = column / 8;
    int columnIndex = column % 8;
    for (int i = 0; i < inBytes * 8; i++) {
        table[i][columnByte] |= ((in[i / 8] & (1 << (i % 8))) >> (i % 8)) << columnIndex; 
    }
}

/*  Flip orientation of table. */
void flipTable(uint8_t **tableOut, const uint8_t **tableIn, int rows, int cols) {
    for (int i = 0; i < rows; i++) {
        for (int j = 0; j < cols; j++) {
            tableOut[j][i/8] |= ((tableIn[i][j / 8] & (1 << (j % 8))) >> (j % 8)) << (i % 8);
        }
    }
}

/* Helper methods. */ 
void convertBignumsToBytes(BIGNUM **src, uint8_t **dst, uint8_t *lens, int sz) {
    for (int i = 0; i < sz; i++) {
        lens[i] = BN_num_bytes(src[i]);
        dst[i] = malloc(lens[i]);
        BN_bn2bin(src[i], dst[i]);
    }
}

BIGNUM *convertBytesToBignum(uint8_t *src, int len) {
    return BN_bin2bn(src, len, NULL);
}

void freeBignums(BIGNUM **bns, int len) {
    for (int i = 0; i < len; i++) {
        BN_free(bns[i]);
    }
}

int getDPFKeyLen() {
    return 2 + 16 + 1 + 18 * LOG_BLOOM_FILTER_SZ + NUM_DOCS_BYTES;
}
int getDPFKeyLen_malicious() {
    return 2 + 16 + 1 + 18 * LOG_BLOOM_FILTER_SZ + (MALICIOUS_DPF_LEN);
}

void printBuffer(char *label, uint8_t *buf, int len) {
    printf("%s: ", label);
    for (int i = 0; i < len; i++) {
        printf("%02x", buf[i]);
    }
    printf("\n");
}

void printByteBinary(char byte) {
    printf(BYTE_TO_BINARY_PATTERN, BYTE_TO_BINARY(byte));
}
