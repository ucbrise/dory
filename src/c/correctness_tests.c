#include "common.h"
#include "server.h"
#include "client.h"
#include "common.h"
#include "params.h"
#include "tokenize.h"
#include <openssl/rand.h>
#include <time.h>
#include <math.h>

int numFailed = 0;

/* Verify updates done correctly. */
int checkUpdates(client *c, server *s1, server *s2, int index, char **keywords, int keywordsLen) {
    int rv;
    uint8_t *bf;
    CHECK_A (bf = malloc(BLOOM_FILTER_BYTES));

    CHECK_C (generateEncryptedBloomFilter(c, bf, keywords, keywordsLen, index, NULL));
    printf("Bloom filter for %d updates for doc %d: ", keywordsLen, index);
    for (int i = 0; i < BLOOM_FILTER_BYTES; i++) {
        printf("%x ", bf[i]);
    }
    printf("\n");


    for (int i = 0; i < BLOOM_FILTER_BYTES; i++) {
        if (s1->indexList[index][i] != bf[i] || s2->indexList[index][i] != bf[i]) {
            printf("FAIL: bloom filter incorrect at byte %d, expected %x but got %x and %x\n", i, bf[i], s1->indexList[index][i], s2->indexList[index][i]);
            return ERROR;
        }
    }
    printf("PASS: Updates for doc %d with %d keywords completed\n", index, keywordsLen);

cleanup:
    if (bf) free(bf);
    return rv;
}

/* Verify  search results correct. */
int checkSearchResults(uint8_t docsPresent[NUM_DOCS_BYTES], int expectedDoc) {
    for (int i = 0; i < NUM_DOCS; i++) {
        // check each bit for 0 or 1
        if (isBitOne(docsPresent, i)) {
            if (i == expectedDoc) {
                printf("PASS: found keyword inserted at doc %d\n", i);
            } else {
                printf("False positive at %d\n", i);
            }
        } else {
            if (i == expectedDoc) {
                printf("FAIL: didn't find keyword inserted at doc %d: 0x%x\n", i, docsPresent[i/8]);
                return ERROR;
            }
        }
    }
    return OKAY;
}

/* Check getBitColumn. */
int checkGetBitColumn() {
    uint8_t **table;
    uint8_t *result;
    table = malloc(MAX_DOCS * sizeof(uint8_t *));
    for (int i = 0; i < MAX_DOCS; i++) {
        table[i] = malloc(1);
        table[i][0] = 0xfe;
    }
    result = malloc(MAX_DOCS_BYTES);
    getBitColumn(table, 0, result);
    for (int i = 0; i < MAX_DOCS_BYTES; i++) {
        if (result[i] != 0) {
            printf("FAIL: expected 0 and got %x\n", result[i]);
            return ERROR;
        }
    }

    getBitColumn(table, 1, result);
    for (int i = 0; i < MAX_DOCS_BYTES; i++) {
        if (result[i] != 0xff) {
            printf("FAIL: expected 0xff and got 0x%x\n", result[i]);
            return ERROR;
        }
    }

    printf("PASS: passed check get bit column tests\n");

    for (int i = 0; i < MAX_DOCS; i++) {
        free(table[i]);
    }
    return OKAY;
}

/* Check flipTable. */
int checkFlipTable() {
    uint8_t **table;
    uint8_t **tableFlipped;
    table = malloc(8 * sizeof(uint8_t *));
    for (int i = 0; i < 8; i++) {
        table[i] = malloc(4);
        table[i][0] = i % 3 == 0 ? 1 : 0;
    }
    uint8_t x = 0xff;
    setBitColumn(table, 1, &x, 1);
    tableFlipped = malloc(32 * sizeof(uint8_t *));
    for (int i = 0; i < 32; i++) {
        tableFlipped[i] = malloc(1);
    }

    printf("before flip\n");
    flipTable(tableFlipped, table, 8, 32);
    printf("after flip\n");

    printf("table: ");
    for (int i = 0; i < 8; i++) {
        for (int j = 0; j < 4; j++) {
            printByteBinary(table[i][j]);
        }
        printf("\n");
    }

    printf("table flipped: ");
    for (int i = 0; i < 32; i++) {
        printByteBinary(tableFlipped[i][0]);
        printf("\n");
    }

    flipTable(table, tableFlipped, 32, 8);
    printf("should be original table: ");
    for (int i = 0; i < 8; i++) {
        for (int j = 0; j < 4; j++) {
            printByteBinary(table[i][j]);
        }
        printf("\n");
    }



}

/* Run tests for semihonest adversary. */
int runSemiHonestTests() {
    int rv;
    client c;
    server s1, s2;
    char ***keywords;
    uint8_t docsPresent[MAX_DOCS_BYTES];

    CHECK_C (initializeClient(&c, 1, NULL, NULL));
    CHECK_C (initializeServer(&s1, 1));
    CHECK_C (initializeServer(&s2, 1));
    printf("Setup complete\n");

    CHECK_A (keywords = malloc(MAX_DOCS * sizeof(char **)));
    for (int i = 0; i < MAX_DOCS; i++) {
        CHECK_A (keywords[i] = malloc(MAX_NUM_KEYWORDS * sizeof(char *)));
        int numKeywords = makeRandKeywords(keywords[i]);
        CHECK_C (runUpdates(&c, &s1, &s2, i, keywords[i], numKeywords));
    }

    printf("done with all updates\n");

    for (int i = 0; i < MAX_DOCS; i++) {
        memset(docsPresent, 0, MAX_DOCS_BYTES);
        CHECK_C (runSearch(&c, &s1, &s2, keywords[i][0], docsPresent));
        if (checkSearchResults(docsPresent, i) == ERROR) {
            numFailed++;
        }
        free(keywords[i]);
    }
    if (checkGetBitColumn() == ERROR) {
        numFailed++;
    }

    freeClient(&c);
    freeServer(&s1);
    freeServer (&s2);

cleanup:
    return rv;
}

/* Run tests for malicious adversary. */
int runMaliciousTests() {
    int rv;
    client c;
    server s1, s2;
    char ***keywords;
    uint8_t docsPresent[NUM_DOCS_BYTES];
    printf("starting malicious tests\n");

    CHECK_C (initializeClient(&c, 1, NULL, NULL));
    CHECK_C (initializeServer(&s1, 1));
    CHECK_C (initializeServer(&s2, 1));
    printf("Setup complete\n");

    CHECK_A (keywords = malloc(NUM_DOCS * sizeof(char **)));
    for (int i = 0; i < NUM_DOCS; i++) {
        CHECK_A (keywords[i] = malloc(MAX_NUM_KEYWORDS * sizeof(char *)));
        int numKeywords = makeRandKeywords(keywords[i]);
        CHECK_C (runUpdates_malicious(&c, &s1, &s2, i, keywords[i], numKeywords));
    }

    printf("done with all updates\n");

    for (int i = 0; i < 1; i++) {
        memset(docsPresent, 0, NUM_DOCS_BYTES);
        CHECK_C (runSearch_malicious(&c, &s1, &s2, keywords[i][0], docsPresent));
        if (checkSearchResults(docsPresent, i) == ERROR) {
            numFailed++;
        }
        free(keywords[i]);
    }
    
    if (checkGetBitColumn() == ERROR) {
        numFailed++;
    }

    freeClient(&c);
    freeServer(&s1);
    freeServer (&s2);

cleanup:
    return rv;
}

int main(int argc, const char *argv[]) {
    int rv;
    NUM_DOCS = 0;
    NUM_DOCS_BYTES = 0;
    MAX_DOCS = 64;
    MAX_DOCS_BYTES = MAX_DOCS / 8;
    BLOOM_FILTER_SZ = 64;
    BLOOM_FILTER_BYTES = BLOOM_FILTER_SZ / 8;
    LOG_BLOOM_FILTER_SZ = 6;

    CHECK_C (runSemiHonestTests());
    CHECK_C (runMaliciousTests());
cleanup:
    if (rv == ERROR) {
        printf("\n*** EXITED EARLY DUE TO ERROR! ***\n");
    } else {
        printf("\n*** %d TESTS FAILED ***\n", numFailed);
    }
}


