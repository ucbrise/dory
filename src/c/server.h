#ifndef _SERVER
#define _SERVER

#include "common.h"
#include <openssl/evp.h>
#include <openssl/bn.h>
#include "params.h"

typedef struct {
    EVP_CIPHER_CTX *ctx[MAX_THREADS];
    uint8_t **indexList;
    uint128_t *macSums;
} server;

int initializeServer(server *s, int numThreads);
void copyServer(server *sDst, server *sSrc);
void freeServer(server *s);
void printServer(server *s);

void setRow(server *s, int i, uint8_t *bf);
int runQuery(server *s, unsigned char *keys[], uint8_t **results, int threadNum, int startIndex, int endIndex);

int setRow_malicious(server *s, int i, uint8_t *bf, uint128_t *macs);
int runQuery_malicious(server *s, unsigned char *keys[], uint8_t **results, int threadNum, int startIndex, int endIndex);
int assemblePerThreadResults(server *s, uint8_t ***in, int numThreads, uint8_t **out);

#endif
