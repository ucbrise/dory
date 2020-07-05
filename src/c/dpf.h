// From SabaEskandarian/OlivKeyValCode

#ifndef _DPF
#define _DPF

#include <stdio.h>
#include <string.h>
#include <stdint.h>

#include <openssl/conf.h>
#include <openssl/evp.h>
#include <openssl/err.h>
#include <string.h>

#include "params.h"

void print_block(uint128_t input);

uint128_t getRandomBlock(void);

//DPF functions

void dpfPRG(EVP_CIPHER_CTX *ctx, uint128_t input, uint128_t* output1, uint128_t* output2, int* bit1, int* bit2);

void genDPF(EVP_CIPHER_CTX *ctx, int domainSize, uint128_t index, int dataSize, uint8_t* data, unsigned char** k0, unsigned char **k1);

uint128_t evalDPF(EVP_CIPHER_CTX *ctx, int domainSize, unsigned char* k, uint128_t x, int dataSize, uint8_t* dataShare);
void evalAllDPF(EVP_CIPHER_CTX *ctx, int domainSize, unsigned char* k, int dataSize, uint8_t **dataShare);

#endif
