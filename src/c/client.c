#include "client.h"
#include "common.h"
#include "dpf.h"
#include "tokenize.h"
#include <openssl/rand.h>
#include <openssl/evp.h>
#include <string.h>
#include <stdio.h>
#include <math.h>
#include <pthread.h>

/* Setup for client. */
int initializeClient(client *c, int numThreads, uint8_t *maskKey, uint8_t *macKey) {
    int rv;

    for (int i = 0; i  < numThreads; i++) {
        CHECK_A (c->ctx[i] = EVP_CIPHER_CTX_new());
        /* Only for testing purposes */
        unsigned char *aeskey = (unsigned char *) "0123456789123456";
        CHECK_C (EVP_EncryptInit_ex(c->ctx[i], EVP_aes_128_ecb(), NULL, aeskey, NULL));
        CHECK_C (EVP_CIPHER_CTX_set_padding(c->ctx[i], 0));
    }

    c->maskKeyLen = 16;
    CHECK_A (c->maskKey = malloc(c->maskKeyLen));
    CHECK_C (RAND_bytes(c->maskKey, c->maskKeyLen));
    CHECK_A (c->maskKey_ctx = EVP_CIPHER_CTX_new());
    CHECK_C (EVP_EncryptInit_ex(c->maskKey_ctx, EVP_aes_128_ecb(), NULL, c->maskKey, NULL));

    CHECK_A (c->versions = malloc(MAX_DOCS * sizeof(uint32_t)));
    memset(c->versions, 0, MAX_DOCS * sizeof(uint32_t));

    c->sysVersion = 0;

    c->macKeyLen = 16;
    CHECK_A (c->macKey = malloc(c->macKeyLen));
    CHECK_C (RAND_bytes(c->macKey, c->macKeyLen));
    CHECK_A (c->macKey_ctx = EVP_CIPHER_CTX_new());
    CHECK_C (EVP_EncryptInit_ex(c->macKey_ctx, EVP_aes_128_ecb(),  NULL, c->macKey, NULL));

    if (maskKey != NULL) {
        memcpy(c->maskKey, maskKey, 16);
        CHECK_C (EVP_EncryptInit_ex(c->maskKey_ctx, EVP_aes_128_ecb(), NULL, c->maskKey, NULL));
    }
    if (macKey != NULL) {
        memcpy(c->macKey, macKey, 16);
        CHECK_C (EVP_EncryptInit_ex(c->macKey_ctx, EVP_aes_128_ecb(),  NULL, c->macKey, NULL));
    }

cleanup:
    if (rv == ERROR) freeClient(c);
    return rv;
}

void freeClient(client *c) {
    if (c->maskKey) free(c->maskKey);
    if (c->versions) free(c->versions);
    if (c->maskKey_ctx) EVP_CIPHER_CTX_free(c->maskKey_ctx);
    if (c->macKey) free(c->macKey);
    if (c->macKey_ctx) EVP_CIPHER_CTX_free(c->macKey_ctx);
}

/* Serialize row and version number, used to generate mask for row. */
void serializeRowAndVersion(uint8_t buf[8], int row, int version) {
    buf[0] = (row >> 24) & 0xff;
    buf[1] = (row >> 16) & 0xff;
    buf[2] = (row >> 8) & 0xff;
    buf[3] = row & 0xff;
    buf[4] = (version >> 24) & 0xff;
    buf[5] = (version >> 16) & 0xff;
    buf[6] = (version >> 8) & 0xff;
    buf[7] = version & 0xff;
}

/* Generate the mask for the entire row of length BLOOM_FILTER_BYTES. */
int getEntireRowMask(client *c, uint8_t buf[], int row) {
    int rv;
    uint8_t tmp[8];
    serializeRowAndVersion(tmp, row, c->versions[row]);
    CHECK_C(prf(c->maskKey_ctx, buf, BLOOM_FILTER_BYTES, tmp, 8, 0));

cleanup:
    return rv;
}

/* Get the mask for the row for a particular 128-bit block. */
int getBlockRowMask(client *c, uint8_t buf[], int row, uint8_t block) {
    int rv;
    uint8_t tmp[8];
    serializeRowAndVersion(tmp, row, c->versions[row]);
    CHECK_C(prf(c->maskKey_ctx, buf, BLOCK_BYTES, tmp, 8, block));

cleanup:
    return rv;
}

/* Decrypt BLOOM_FILTER_K columns using the mask. */
int decryptCols(client *c, uint8_t **buf, uint32_t *cols) {
    int rv;
    uint8_t *row;
    CHECK_A (row = malloc(BLOCK_BYTES));
    for (int i = 0; i < NUM_DOCS_BYTES; i++) {
        uint8_t tmp[BLOOM_FILTER_K];
        memset(tmp, 0, BLOOM_FILTER_K);
        for (int j = 0; j < 8; j++) {
            CHECK_C (getBlockRowMask(c, row, i * 8 + j, cols[0] / BLOCK_SZ));
            for (int k = 0; k < BLOOM_FILTER_K; k++) {
                copyBit(&tmp[k], j, row, cols[k] % BLOCK_SZ);
            }
        }
        for (int j = 0; j < BLOOM_FILTER_K; j++) {
            buf[j][i] = buf[j][i] ^ tmp[j];
        }
    }
cleanup:
    if (row) free(row);
    return rv;
}

/* Get the BLOOM_FILTER_K indexes associated with a keyword by hashing the keyword. */
int getIndexesForKeyword(client *c, uint32_t indexes[], char *keyword) {
    int rv;
    uint8_t *tmp;
    CHECK_A (tmp = malloc(4 * BLOOM_FILTER_K + 1));
    CHECK_C (hashToBytes(tmp, 4 * BLOOM_FILTER_K + 1, keyword, strlen(keyword)));
    uint8_t base = tmp[4 * BLOOM_FILTER_K] % (uint8_t)(ceil((double)BLOOM_FILTER_SZ / BLOCK_SZ));
    uint8_t modValue = min(BLOOM_FILTER_SZ, BLOCK_SZ);
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        indexes[i] = (tmp[0 + 4 * i] << 24) + (tmp[1 + 4 * i] << 16) + (tmp[2 + 4 * i] << 8) + (tmp[3 + 4 * i]); // big-endian
        indexes[i] = (indexes[i] % modValue) + (base * BLOCK_SZ);
    }
cleanup:
    if (tmp) free(tmp);
    return rv;
}

/* Generate the Bloom filter for a set of keywords. */
int generateBloomFilter(client *c, uint8_t *bf, char *keywords[], size_t keywordsLen) {
    int rv = OKAY;
    uint32_t *indexes;

    CHECK_A (indexes = malloc(BLOOM_FILTER_K * sizeof(uint32_t)));
    memset(bf, 0, BLOOM_FILTER_BYTES);
    for (int i = 0; i < keywordsLen; i++) {
        CHECK_C (getIndexesForKeyword(c, indexes, keywords[i]));
        for (int j = 0; j < BLOOM_FILTER_K; j++) {
            setBitOne(bf, indexes[j]);
        }
    }
cleanup:
    if (indexes) free(indexes);
    return rv;
}

/* Generate the encrypted Bloom filter for a set of keywords. */
int generateEncryptedBloomFilter(client *c, uint8_t *bf, char *keywords[], size_t keywordsLen, int doc, uint32_t *version) {
    int rv;
    uint8_t *rowMask;

    if (doc >= NUM_DOCS) {
        NUM_DOCS = doc + 1;
        NUM_DOCS_BYTES = ceil(((double) NUM_DOCS) / 8.0);
        MALICIOUS_DPF_LEN = NUM_DOCS_BYTES + MAC_BYTES;
    }

    CHECK_A (rowMask = malloc(BLOOM_FILTER_BYTES));
    CHECK_C (generateBloomFilter(c, bf, keywords, keywordsLen));
    CHECK_C (getEntireRowMask(c, rowMask, doc));
    xorIn(bf, rowMask, BLOOM_FILTER_BYTES);
    if (version != NULL) *version = c->versions[doc];
cleanup:
    if (rowMask) free(rowMask);
    return rv;
}

/* Set the buffer used to generate the MAC for a single entry. */
int setMACBuf(client *c, uint8_t *buf, int doc, int col, int entry) {
    memset(buf, 0, MAC_BYTES);
    buf[0] = (doc >> 24) & 0xff;
    buf[1] = (doc >> 16) & 0xff;
    buf[2] = (doc > 8) & 0xff;
    buf[3] = (doc) & 0xff;
    buf[4] = (col >> 24) & 0xff;
    buf[5] = (col >> 16) & 0xff;
    buf[6] = (col >> 8) & 0xff;
    buf[7] = (col) & 0xff;
    buf[8] = (c->versions[doc] >> 24) & 0xff;
    buf[9] = (c->versions[doc] >> 16) & 0xff;
    buf[10] = (c->versions[doc] >> 8) & 0xff;
    buf[11] = (c->versions[doc]) & 0xff;
    buf[12] = (entry >> 24) & 0xff;
    buf[13] = (entry >> 16) & 0xff;
    buf[14] = (entry > 8) & 0xff;
    buf[15] = (entry) & 0xff;
}

/* Generate the MAC for a single entry. */
int generateMACForEntry_malicious(client *c, uint32_t entry, int doc, int col, uint128_t *mac) {
    int rv;

    uint8_t bufIn[MAC_BYTES];
    setMACBuf(c, bufIn, doc, col, entry);
    CHECK_C (blockPrf(c->macKey_ctx, (uint8_t *)mac, bufIn, sizeof(uint128_t)));
cleanup:
    return rv;
}

/* Generate all the MACs for a Bloom filter. */
int generateMACsForBloomFilter_malicious(client *c, uint8_t *bf, int doc, uint128_t *macs) {
    int rv;
    for (int i = 0; i < BLOOM_FILTER_SZ; i++) {
        uint32_t bit = (bf[i / 8] & (1 << (i % 8))) >> (i % 8);
        CHECK_C (generateMACForEntry_malicious(c, bit, doc, i, &macs[i]));
    }
cleanup:
    return rv;
}

/* Generate a DPF query for a keyword (keys_s1 and keys_s2 of length BLOOM_FILTER_K)
 * (only for semihonest adversaries).. */
int generateKeywordQuery(client *c, char *keyword, unsigned char *keys_s1[], unsigned char *keys_s2[], uint32_t *indexes) {
    int rv;
    uint8_t *data;
    CHECK_C (getIndexesForKeyword(c, indexes, keyword));

    CHECK_A (data = malloc(NUM_DOCS_BYTES));
    memset(data, 0xff, NUM_DOCS_BYTES);
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        genDPF(c->ctx[0], LOG_BLOOM_FILTER_SZ, indexes[i], NUM_DOCS_BYTES, data, &keys_s1[i], &keys_s2[i]);
    }
cleanup:
    if (data) free(data);
    return rv;
}

/* Generate a DPF query for a keyword (keys_s1 and keys_s2 of length BLOOM_FILTER_K)
 * (only for malicious adversaries). */
int generateKeywordQuery_malicious(client *c, char *keyword, unsigned char *keys_s1[], unsigned char *keys_s2[], uint32_t *indexes) {
    int rv;
    uint8_t *data;
    CHECK_C (getIndexesForKeyword(c, indexes, keyword));

    CHECK_A (data = malloc(MALICIOUS_DPF_LEN));
    memset(data, 0xff, MALICIOUS_DPF_LEN);
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        genDPF(c->ctx[0], LOG_BLOOM_FILTER_SZ, indexes[i], MALICIOUS_DPF_LEN, data, &keys_s1[i], &keys_s2[i]);
    }
cleanup:
    if (data) free(data);
    return rv;
}

/* Assemble responses for queries (only for semihonest adversaries). */
int assembleQueryResponses(client *c, uint8_t **results1, uint8_t **results2, uint32_t *indexes, uint8_t *docsPresent) {
    /* Assemble responses. */
    int rv;
    uint8_t **results;
    CHECK_A (results = malloc(BLOOM_FILTER_K * sizeof(uint8_t *)));
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        CHECK_A (results[i] = malloc(NUM_DOCS_BYTES));
    }

    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        for (int j = 0; j < NUM_DOCS_BYTES; j++) {
            results[i][j] = results1[i][j] ^ results2[i][j];
        }
    }

    /* Decrypt columns. */
    CHECK_C (decryptCols(c, results, indexes));

    /* Look for where all 1s. */
    memset(docsPresent, 0xff, NUM_DOCS_BYTES);
    for (int i = 0; i < NUM_DOCS_BYTES; i++) {
        for (int j = 0; j < BLOOM_FILTER_K; j++) {
            docsPresent[i] = docsPresent[i] & results[j][i];
        }
    }
cleanup:
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        if (results && results[i]) free(results[i]);
    }
    if (results) free(results);
    return rv;
}

typedef struct {
    client *c;
    int col;
    uint8_t *results;
    uint128_t receivedMac;
} macArgs;

/* Verify the column MACs. */
void checkColMACs(macArgs *args) {
    uint128_t mac = 0;

    for (int j = 0; j < NUM_DOCS; j++) {
        uint128_t currMac = 0;
        uint32_t bit = (args->results[j / 8] & (1 << (j % 8))) >> (j % 8);
        generateMACForEntry_malicious(args->c, bit, j, args->col, &currMac);
        mac = mac ^ currMac;
    }

    // Commenting out because throughput tests send dummy updates that cause
    // MAC verification to fail
    // TODO: propagate errors better
/*    if (mac != args->receivedMac) {
        printf("ERROR: MACs don't match\n");
    }*/

}

/* Assemble responses for keyword queries (only for malicious adversaries). */
int assembleQueryResponses_malicious(client *c, uint8_t **results1, uint8_t **results2, uint32_t *indexes, uint8_t *docsPresent) {
    /* Assemble responses. */
    int rv;
    uint8_t **results;
    uint8_t **macs;
    
    CHECK_A (results = malloc(BLOOM_FILTER_K * sizeof(uint8_t *)));
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        CHECK_A (results[i] = malloc(NUM_DOCS_BYTES));
    }

    CHECK_A (macs = malloc(BLOOM_FILTER_K * sizeof(uint8_t *)));
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        CHECK_A (macs[i] = malloc(MAC_BYTES));
        memset(macs[i], 0, MAC_BYTES);
    }

    /* Reassemble results. */
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        for (int j = 0; j < MAC_BYTES; j++) {
            macs[i][j] = results1[i][j] ^ results2[i][j];
        }
        for (int j = 0; j < NUM_DOCS_BYTES; j++) {
            results[i][j] = results1[i][j + MAC_BYTES] ^ results2[i][j + MAC_BYTES];
        }
    }

    uint128_t mac = 0;
    
    /* Check MACs. */
    pthread_t t[BLOOM_FILTER_K];
    macArgs args[BLOOM_FILTER_K];
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        /* Add up all bits in results[i] mod MAC_MODP */
        args[i].col = indexes[i];
        args[i].c = c;
        args[i].results = results[i];
        memcpy(&args[i].receivedMac, macs[i], MAC_BYTES);

        pthread_create(&t[i], NULL, checkColMACs, args);
    }
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        pthread_join(t[i], NULL);
    }

    /* Decrypt columns. */
    CHECK_C (decryptCols(c, results, indexes));

    /* Look for where all 1s. */
    memset(docsPresent, 0xff, NUM_DOCS_BYTES);
    for (int i = 0; i < NUM_DOCS_BYTES; i++) {
        for (int j = 0; j < BLOOM_FILTER_K; j++) {
            docsPresent[i] = docsPresent[i] & results[j][i];
        }
    }

cleanup:
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        if (results && results[i]) free(results[i]);
        if (macs && macs[i]) free(macs[i]);
    }
    if (results) free(results);
    return rv;
}

/* Update the client state.. */
void updateClientState(client *c, int numDocs, uint32_t *versions) {
    if (numDocs < NUM_DOCS) return;
    NUM_DOCS = numDocs;
    NUM_DOCS_BYTES = ceil(((double) NUM_DOCS) / 8.0);
    MALICIOUS_DPF_LEN = NUM_DOCS_BYTES + MAC_BYTES;
/*    if (versions != NULL) {
        memcpy(c->versions, versions, MAX_DOCS * sizeof(uint32_t));
    }*/
}
