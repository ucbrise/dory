#include "server.h"
#include "common.h"
#include <openssl/rand.h>
#include <string.h>
#include <math.h>

/* Setup for server. */
int initializeServer(server *s, int numThreads) {
    int rv; 

    for (int i = 0; i  < numThreads; i++) {
        CHECK_A (s->ctx[i] = EVP_CIPHER_CTX_new());
        /* Only for testing. */
        unsigned char *aeskey = (unsigned char *) "0123456789123456";
        CHECK_C (EVP_EncryptInit_ex(s->ctx[i], EVP_aes_128_ecb(), NULL, aeskey, NULL));
        CHECK_C (EVP_CIPHER_CTX_set_padding(s->ctx[i], 0));
    }

    CHECK_A (s->indexList = malloc(BLOOM_FILTER_SZ * sizeof(uint8_t *)));
    for (int i = 0; i < BLOOM_FILTER_SZ; i++) {
        CHECK_A (s->indexList[i] = malloc(MAX_DOCS_BYTES));
        memset(s->indexList[i], 0, MAX_DOCS_BYTES);
    } 

    CHECK_A (s->macSums = malloc(BLOOM_FILTER_SZ * sizeof(uint128_t)));
    for (int i = 0; i < BLOOM_FILTER_SZ; i++) {
        s->macSums[i] = 0;
    }

cleanup:
    if (rv == ERROR) freeServer(s);
    return rv; 
}

/* Copy state of server. */
void copyServer(server *sDst, server *sSrc) {
    for (int i = 0; i < BLOOM_FILTER_SZ; i++) {
        memcpy(sDst->indexList[i], sSrc->indexList[i], MAX_DOCS_BYTES);
        sDst->macSums[i] = sSrc->macSums[i];
    }
}

/* Print state of server. */
void printServer(server *s) {
    printf("starting print server\n");
    if (s == NULL) printf("s is null\n");
    for (int i = 0; i < BLOOM_FILTER_SZ; i++) {
        printf("row %d:", i);
        printBuffer("", s->indexList[i], MAX_DOCS_BYTES);
    }
    for (int i = 0; i < BLOOM_FILTER_SZ; i++) {
        printf("aggregate MAC %d: %x\n", i, s->macSums[i]);
    }
}

/* Free state of server. */
void freeServer(server *s) {
    for (int i = 0; i < BLOOM_FILTER_SZ; i++) {
        if (s->indexList && s->indexList[i]) free(s->indexList[i]);
    }
    if (s->indexList) free(s->indexList);

    free(s->macSums);
}

/* Set a row in the table (update for a document) (semihonest adversaries).. */
void setRow(server *s, int i, uint8_t *bf) {
    if (i >= NUM_DOCS) {
        NUM_DOCS = i + 1;
        NUM_DOCS_BYTES = ceil(((double) NUM_DOCS) / 8.0);
        MALICIOUS_DPF_LEN = NUM_DOCS_BYTES + MAC_BYTES;
    }
    setBitColumn(s->indexList, i, bf, BLOOM_FILTER_BYTES);
}

/* Set a row in the table (update for a document) (malicious adversaries). */
int setRow_malicious(server *s, int i, uint8_t *bf, uint128_t *macs) {
    int rv;
    setRow(s, i, bf);
    for (int j = 0; j < BLOOM_FILTER_SZ; j++) {
        if (s->macSums[j] != macs[j]) {
            s->macSums[j] = s->macSums[j] ^ macs[j];
        }
    }
cleanup:
    return rv;
}

/* Execute query for semihonest adversaries. */
int runQuery_leaky(server *s, uint32_t *indexes, uint8_t **results) {
    int rv;
    uint8_t *output;

    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        if (indexes[i] >= BLOOM_FILTER_SZ) printf("OUT OF BOUNDS: %d / %d\n", indexes[i], BLOOM_FILTER_SZ);
        memcpy(results[i], s->indexList[indexes[i]], NUM_DOCS_BYTES);
    }
    
cleanup:
    return rv;
}



/* Execute query for semihonest adversaries. */
int runQuery(server *s, unsigned char *keys[], uint8_t **results, int threadNum, int startIndex, int endIndex) {
    int rv;
    uint8_t *output;

    int outputLen = NUM_DOCS_BYTES + (16 - (NUM_DOCS_BYTES) % 16);
    CHECK_A (output = malloc(outputLen));
    memset(output, 0, outputLen);

    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        memset(results[i], 0, NUM_DOCS_BYTES);
    }
    
    for (uint128_t j = startIndex; j < endIndex; j++) {
        for (int i = 0; i < BLOOM_FILTER_K; i++) {
            evalDPF(s->ctx[threadNum], LOG_BLOOM_FILTER_SZ, keys[i], j, outputLen, output);
            for (int k = 0; k < NUM_DOCS_BYTES; k++) {
                results[i][k] = results[i][k] ^ (s->indexList[j][k] & output[k]);
            }
        }
    }
    
cleanup:
    //if (output) free(output);
    return rv;
}

/* Execute query for malicious adversaries. */
int runQuery_malicious(server *s, unsigned char *keys[], uint8_t **results, int threadNum, int startIndex, int endIndex) {
    int rv;
    uint8_t *tmp;
    uint8_t *output;

    if (startIndex < 0 || endIndex > BLOOM_FILTER_SZ ||  startIndex > endIndex) {
        printf("ERROR: bad start index %d and/or end index %d\n", startIndex, endIndex);
    }
    
    int outputLen = MALICIOUS_DPF_LEN + (16 - (MALICIOUS_DPF_LEN) % 16);
    CHECK_A (output = malloc(outputLen));
    memset(output, 0, outputLen);

    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        memset(results[i], 0, MALICIOUS_DPF_LEN);
    }

    for (uint128_t j = startIndex; j < endIndex; j++) {
        for (int i = 0; i < BLOOM_FILTER_K; i++) {
            evalDPF(s->ctx[threadNum], LOG_BLOOM_FILTER_SZ, keys[i], j, outputLen, output);
            for (int k = 0; k < MAC_BYTES; k++) {
                results[i][k] = results[i][k] ^ (((uint8_t *)&s->macSums[j])[k] & output[k]);
            }
            for (int k = 0; k < NUM_DOCS_BYTES; k++) {
                results[i][k + MAC_BYTES] = results[i][k + MAC_BYTES] ^ (s->indexList[j][k] & output[k + MAC_BYTES]);
            }
        }
    }

cleanup:
    if (output) free(output);
    return rv;
}

/* Assemble results generated by individual threads for executing query (malicious adversaries). */
int assemblePerThreadResults_semihonest(server *s, uint8_t ***in, int numThreads, uint8_t **out) {
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        memset(out[i], 0, NUM_DOCS_BYTES);
        for (int j = 0; j < NUM_DOCS_BYTES; j++) {
            for (int k = 0; k < numThreads; k++) {
                out[i][j] = out[i][j] ^ in[k][i][j];
            }
        }
    }
}
/* Assemble results generated by individual threads for executing query (malicious adversaries). */
int assemblePerThreadResults_malicious(server *s, uint8_t ***in, int numThreads, uint8_t **out) {
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        memset(out[i], 0, MALICIOUS_DPF_LEN);
        for (int j = 0; j < MALICIOUS_DPF_LEN; j++) {
            for (int k = 0; k < numThreads; k++) {
                out[i][j] = out[i][j] ^ in[k][i][j];
            }
        }
    }
}
