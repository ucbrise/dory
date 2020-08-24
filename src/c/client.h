// From SabaEskandarian/OlivKeyValCode

#ifndef _CLIENT
#define _CLIENT

#define DO_MAC_CHECK

#include "common.h"
#include <openssl/evp.h>
#include "params.h"
#include <openssl/bn.h>

typedef struct {
    EVP_CIPHER_CTX *ctx[MAX_THREADS];
    uint8_t *indexKey;
    int indexKeyLen;
    uint8_t *maskKey;
    EVP_CIPHER_CTX *maskKey_ctx;
    int maskKeyLen;
    uint32_t *versions;
    uint32_t sysVersion;
    uint8_t *macKey;
    EVP_CIPHER_CTX *macKey_ctx;
    int macKeyLen;
} client;

/* Create and free client state. */
int initializeClient(client *c, int numThreads, uint8_t *maskKey, uint8_t *macKey);
void freeClient(client *c);

/* Generate update. */
int generateBloomFilter(client *c, uint8_t *bf, char *keywords[], size_t keywordsLen);
int generateEncryptedBloomFilter(client *c, uint8_t *bf, char *keywords[], size_t keywordsLen, int doc, uint32_t *version);
int generateMACsForBloomFilter_malicious(client *c, uint8_t *bf, int doc, uint128_t *macs);

/* Perform search (semihonest adversary). */
int generateKeywordQuery(client *c, char *keyword, unsigned char *keys_s1[], unsigned char *keys_s2[], uint32_t *indexes);
int assembleQueryResponses(client *c, uint8_t **results1, uint8_t **results2, uint32_t *indexes,uint8_t *docsPresent);

/* Perform search (malicious adversary). */
int generateKeywordQuery_malicious(client *c, char *keyword, unsigned char *keys_s1[], unsigned char *keys_s2[], uint32_t *indexes);
int assembleQueryResponses_malicious(client *c, uint8_t **results1, uint8_t **results2, uint32_t *indexes, uint8_t *docsPresent);

/* Update state. */
void updateClientState(client *c, int numDocs, uint32_t *versions);

#endif
