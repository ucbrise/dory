#include "common.h"
#include "server.h"
#include "client.h"
#include "common.h"
#include "params.h"
#include <openssl/rand.h>
#include <time.h>
#include <math.h>

int makeRandKeywords(char **keywords) {

    initStemmer();
    int numLines = 10;
    char **lines = malloc(sizeof(char *) * numLines);
    char *strLiteral = "hello world!";
    for (int i = 0; i < numLines; i++) {
        lines[i] = strdup(strLiteral);
    }

    return tokenizeDoc(keywords, lines, numLines);
}

int runUpdates(client *c, server *s1, server *s2, int index, char **keywords, int keywordsLen) {
    int rv;
    uint8_t *bf;
    CHECK_A (bf = malloc(BLOOM_FILTER_BYTES));

    CHECK_C (generateEncryptedBloomFilter(c, bf, keywords, keywordsLen, index, NULL));

    setRow(s1, index, bf);
    setRow(s2, index, bf);

cleanup:
    if (bf) free(bf);
    return rv;
}

int runUpdates_malicious(client *c, server *s1, server *s2, int index, char **keywords, int keywordsLen) {
    int rv;
    uint8_t *bf;
    uint128_t macs[BLOOM_FILTER_SZ];
    CHECK_A (bf = malloc(BLOOM_FILTER_BYTES));

    CHECK_C (generateEncryptedBloomFilter(c, bf, keywords, keywordsLen, index, NULL));
    CHECK_C (generateMACsForBloomFilter_malicious(c, bf, index, macs));

    setRow_malicious(s1, index, bf, macs);
    setRow_malicious(s2, index, bf, macs);

cleanup:
    if (bf) free(bf);
    return rv;
}


int runSearch(client *c, server *s1, server *s2, char *keyword, uint8_t docsPresent[NUM_DOCS_BYTES]) {
    int rv;
    unsigned char **keys_s1;
    unsigned char **keys_s2;
    uint8_t **results1;
    uint8_t **results2;
    uint32_t *indexes;

    CHECK_A (keys_s1 = malloc(BLOOM_FILTER_K * sizeof(unsigned char *)));
    CHECK_A (keys_s2 = malloc(BLOOM_FILTER_K * sizeof(unsigned char *)));
    CHECK_A (results1 = malloc(BLOOM_FILTER_K * sizeof(uint8_t *)));
    CHECK_A (results2 = malloc(BLOOM_FILTER_K * sizeof(uint8_t *)));
    CHECK_A (indexes = malloc(BLOOM_FILTER_K * sizeof(uint32_t)));
    for (int i = 0 ; i < BLOOM_FILTER_K; i++) {
        CHECK_A (results1[i] = malloc(NUM_DOCS_BYTES * sizeof(uint8_t)));
        CHECK_A (results2[i] = malloc(NUM_DOCS_BYTES * sizeof(uint8_t)));
    }

    CHECK_C (generateKeywordQuery(c, keyword, keys_s1, keys_s2, indexes));
    CHECK_C (runQuery(s1, keys_s1, results1, 0, 0, BLOOM_FILTER_SZ));
    CHECK_C (runQuery(s2, keys_s2, results2, 0, 0, BLOOM_FILTER_SZ));
    assembleQueryResponses(c, results1, results2, indexes, docsPresent);
cleanup:
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        if (results1 && results1[i]) free(results1[i]);
        if (results2 && results2[i]) free(results2[i]);
    }
    if (results1) free(results1);
    if (results2) free(results2);
    if (indexes) free(indexes);
    return rv;
}

int runSearch_malicious(client *c, server *s1, server *s2, char *keyword, uint8_t docsPresent[NUM_DOCS_BYTES]) {
    int rv;
    unsigned char *keys_s1[BLOOM_FILTER_K];
    unsigned char *keys_s2[BLOOM_FILTER_K];
    uint8_t **results1;
    uint8_t **results2;
    uint32_t indexes[BLOOM_FILTER_K];
    int numThreads = 2;

    CHECK_A (results1 = malloc(BLOOM_FILTER_K * sizeof(uint8_t *)));
    CHECK_A (results2 = malloc(BLOOM_FILTER_K * sizeof(uint8_t *)));
    for (int i = 0 ; i < BLOOM_FILTER_K; i++) {
        CHECK_A (results1[i] = malloc((MALICIOUS_DPF_LEN)));
        CHECK_A (results2[i] = malloc((MALICIOUS_DPF_LEN)));
    }

    CHECK_C (generateKeywordQuery_malicious(c, keyword, keys_s1, keys_s2, indexes));
    CHECK_C (runQuery_malicious(s1, keys_s1, results1, 0, 0, BLOOM_FILTER_SZ));
    CHECK_C (runQuery_malicious(s2, keys_s2, results2, 0, 0, BLOOM_FILTER_SZ));
    CHECK_C (assembleQueryResponses_malicious(c, results1, results2, indexes, docsPresent));
cleanup:
    for (int i = 0; i < BLOOM_FILTER_K; i++) {
        if (results1 && results1[i]) free(results1[i]);
        if (results2 && results2[i]) free(results2[i]);
    }
    if (results1) free(results1);
    if (results2) free(results2);
    return rv;
}

void printTables(server *s1, server *s2) {
    printf("->SERVER 1\n");
    for (int i = 0; i < NUM_DOCS; i++) {
        for (int j = 0; j < BLOOM_FILTER_BYTES; j++) {
            printf("%x ", s1->indexList[i][j]);
        }
        printf("\n");
    }

    printf("->SERVER 2\n");
    for (int i = 0; i < NUM_DOCS; i++) {
        for (int j = 0; j < BLOOM_FILTER_BYTES; j++) {
            printf("%x ", s2->indexList[i][j]);
        }
        printf("\n");
    }
}

int parseArgs(int argc, const char *argv[]) {
    printf("Parsing arguments...\n");
    if (argc != 3) {
        printf("Expecting MAX_DOCS and BLOOM_FILTER_SZ\n");
        return ERROR;
    }

    NUM_DOCS = 0;
    NUM_DOCS_BYTES = 0;

    MAX_DOCS = atoi(argv[1]);
    if (MAX_DOCS <= 0 || MAX_DOCS % 8 != 0) {
        printf("Bad MAX_DOCS value: %d\n", NUM_DOCS);
        return ERROR;
    }
    MAX_DOCS_BYTES = MAX_DOCS / 8;

    BLOOM_FILTER_SZ = atoi(argv[2]);
    if (BLOOM_FILTER_SZ <= 0 || BLOOM_FILTER_SZ % 8 != 0) {
        printf("Bad BLOOM_FILTER_SZ: %d\n", BLOOM_FILTER_SZ);
        return ERROR;
    }
    BLOOM_FILTER_BYTES = BLOOM_FILTER_SZ / 8;

    LOG_BLOOM_FILTER_SZ = ((int) log2((double)BLOOM_FILTER_SZ));

    printf("MAX_DOCS = %d, BLOOM_FILTER_SZ = %d\n", MAX_DOCS, BLOOM_FILTER_SZ);

    return OKAY;
}
