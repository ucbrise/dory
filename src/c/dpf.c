// From SabaEskandarian/OblivKeyValCode

#include "dpf.h"
#include <openssl/rand.h>
//#include <omp.h>
#include <time.h>
#include "params.h"
#include <math.h>

uint128_t dpf_reverse_lsb(uint128_t input){
    uint128_t xor = 1;
	return input ^ xor;
}

uint128_t dpf_lsb(uint128_t input){
    return input & 1;
}

uint128_t dpf_set_lsb_zero(uint128_t input){
    int lsb = input & 1;

	if(lsb == 1){
		return dpf_reverse_lsb(input);
	}else{
		return input;
	}
}

void _output_bit_to_bit(uint128_t input){
    for(int i = 0; i < 64; i++)
    {
        if( (1ll << i) & input)
            printf("1");
	else
	    printf("0");
    }
}

void print_block(uint128_t input) {
    uint64_t *val = (uint64_t *) &input;

	_output_bit_to_bit(val[0]);
	_output_bit_to_bit(val[1]);
	printf("\n");
}

uint128_t getRandomBlock(){
    static uint8_t* randKey = NULL;//(uint8_t*) malloc(16);
    static EVP_CIPHER_CTX* randCtx;
    static uint128_t counter = 0;

    int len = 0;
    uint128_t output = 0;
    if(!randKey){
        randKey = (uint8_t*) malloc(16);
        if(!(randCtx = EVP_CIPHER_CTX_new()))
            printf("errors occured in creating context\n");
        if(!RAND_bytes(randKey, 16)){
            printf("failed to seed randomness\n");
        }
        if(1 != EVP_EncryptInit_ex(randCtx, EVP_aes_128_ecb(), NULL, randKey, NULL))
            printf("errors occured in randomness init\n");
        EVP_CIPHER_CTX_set_padding(randCtx, 0);
    }

    if(1 != EVP_EncryptUpdate(randCtx, (uint8_t*)&output, &len, (uint8_t*)&counter, 16))
        printf("errors occured in generating randomness\n");
    counter++;
    return output;
}

//this is the PRG used for the DPF
void dpfPRG(EVP_CIPHER_CTX *ctx, uint128_t input, uint128_t* output1, uint128_t* output2, int* bit1, int* bit2){

	input = dpf_set_lsb_zero(input);

    int len = 0;
	uint128_t stashin[2];
	stashin[0] = input;
	stashin[1] = dpf_reverse_lsb(input);
	uint128_t stash[2];

    EVP_CIPHER_CTX_set_padding(ctx, 0);
    if(1 != EVP_EncryptUpdate(ctx, (uint8_t*)stash, &len, (uint8_t*)stashin, 32))
        printf("errors occured in encrypt\n");
    //no need to do this since we're working with exact multiples of the block size
    //if(1 != EVP_EncryptFinal_ex(ctx, stash + len, &len))
    //    printf("errors occured in final\n");

	stash[0] = stash[0] ^ input;
	stash[1] = stash[1] ^ input;
	stash[1] = dpf_reverse_lsb(stash[1]);

	*bit1 = dpf_lsb(stash[0]);
	*bit2 = dpf_lsb(stash[1]);

	*output1 = dpf_set_lsb_zero(stash[0]);
	*output2 = dpf_set_lsb_zero(stash[1]);
}

static int getbit(uint128_t x, int n, int b){
	return ((uint128_t)(x) >> (n - b)) & 1;
}

void genDPF(EVP_CIPHER_CTX *ctx, int domainSize, uint128_t index, int dataSize, uint8_t* data, unsigned char** k0, unsigned char **k1){
    int maxLayer = domainSize;

    uint128_t s[maxLayer + 1][2];
	int t[maxLayer + 1 ][2];
	uint128_t sCW[maxLayer];
	int tCW[maxLayer][2];

    s[0][0] = getRandomBlock();
	s[0][1] = getRandomBlock();
	t[0][0] = 0;
	t[0][1] = 1;

    uint128_t s0[2], s1[2]; // 0=L,1=R
    int t0[2], t1[2];
	#define LEFT 0
	#define RIGHT 1
	for(int i = 1; i <= maxLayer; i++){
        dpfPRG(ctx, s[i-1][0], &s0[LEFT], &s0[RIGHT], &t0[LEFT], &t0[RIGHT]);
		dpfPRG(ctx, s[i-1][1], &s1[LEFT], &s1[RIGHT], &t1[LEFT], &t1[RIGHT]);

        int keep, lose;
		int indexBit = getbit(index, domainSize, i);
        if(indexBit == 0){
			keep = LEFT;
			lose = RIGHT;
		}else{
			keep = RIGHT;
			lose = LEFT;
		}


        sCW[i-1] = s0[lose] ^ s1[lose];

		tCW[i-1][LEFT] = t0[LEFT] ^ t1[LEFT] ^ indexBit ^ 1;
		tCW[i-1][RIGHT] = t0[RIGHT] ^ t1[RIGHT] ^ indexBit;

		if(t[i-1][0] == 1){
			s[i][0] = s0[keep] ^ sCW[i-1];
			t[i][0] = t0[keep] ^ tCW[i-1][keep];
		}else{
			s[i][0] = s0[keep];
			t[i][0] = t0[keep];
		}

		if(t[i-1][1] == 1){
			s[i][1] = s1[keep] ^ sCW[i-1];
			t[i][1] = t1[keep] ^ tCW[i-1][keep];
		}else{
			s[i][1] = s1[keep];
			t[i][1] = t1[keep];
		}

    }

	unsigned char *buff0;
	unsigned char *buff1;
	buff0 = (unsigned char*) malloc(1 + 16 + 1 + 18 * maxLayer + dataSize);
	buff1 = (unsigned char*) malloc(1 + 16 + 1 + 18 * maxLayer + dataSize);

    //take data, xor it with dataSize bits generated from s_0^n and another dataSize bits generated from s_1^n
    //use a counter mode encryption of 0 with each seed as key to get prg output
    uint8_t *lastCW = (uint8_t*) malloc(dataSize);
    uint8_t *convert0 = (uint8_t*) malloc(dataSize+16);
    uint8_t *convert1 = (uint8_t*) malloc(dataSize+16);
    uint8_t *zeros = (uint8_t*) malloc(dataSize+16);
    memset(zeros, 0, dataSize+16);

    memcpy(lastCW, data, dataSize);

    int len = 0;
    //generate dataSize length prg outputs from the seeds

    EVP_CIPHER_CTX *seedCtx0;
    EVP_CIPHER_CTX *seedCtx1;


    if(!(seedCtx0 = EVP_CIPHER_CTX_new()))
        printf("errors occured in creating context\n");
    if(!(seedCtx1 = EVP_CIPHER_CTX_new()))
        printf("errors occured in creating context\n");
    if(1 != EVP_EncryptInit_ex(seedCtx0, EVP_aes_128_ctr(), NULL, (uint8_t*)&s[maxLayer][0], NULL))
        printf("errors occured in init of dpf gen\n");
    if(1 != EVP_EncryptInit_ex(seedCtx1, EVP_aes_128_ctr(), NULL, (uint8_t*)&s[maxLayer][1], NULL))
        printf("errors occured in init of dpf gen\n");
    if(1 != EVP_EncryptUpdate(seedCtx0, convert0, &len, zeros, dataSize))
        printf("errors occured in encrypt\n");
    if(1 != EVP_EncryptUpdate(seedCtx1, convert1, &len, zeros, dataSize))
        printf("errors occured in encrypt\n");

    for(int i = 0; i < dataSize; i++){
        lastCW[i] = lastCW[i] ^ ((uint8_t*)convert0)[i] ^ ((uint8_t*)convert1)[i];
    }

	if(buff0 == NULL || buff1 == NULL){
		printf("Memory allocation failed\n");
		exit(1);
	}

	buff0[0] = domainSize;
	memcpy(&buff0[1], &s[0][0], 16);
	buff0[17] = t[0][0];
	for(int i = 1; i <= maxLayer; i++){
		memcpy(&buff0[18 * i], &sCW[i-1], 16);
		buff0[18 * i + 16] = tCW[i-1][0];
		buff0[18 * i + 17] = tCW[i-1][1];
	}
	memcpy(&buff0[18 * maxLayer + 18], lastCW, dataSize);

	buff1[0] = domainSize;
	memcpy(&buff1[18], &buff0[18], 18 * (maxLayer));
	memcpy(&buff1[1], &s[0][1], 16);
	buff1[17] = t[0][1];
	memcpy(&buff1[18 * maxLayer + 18], lastCW, dataSize);

	*k0 = buff0;
	*k1 = buff1;

    free(lastCW);
    free(convert0);
    free(convert1);
    free(zeros);
    EVP_CIPHER_CTX_free(seedCtx0);
    EVP_CIPHER_CTX_free(seedCtx1);
}

uint128_t evalDPF(EVP_CIPHER_CTX *ctx, int domainSize, unsigned char* k, uint128_t x, int dataSize, uint8_t* dataShare){

    //dataShare is of size dataSize
 	int n = domainSize;
	int maxLayer = n;

	uint128_t s[maxLayer + 1];
	int t[maxLayer + 1];
	uint128_t sCW[maxLayer];
	int tCW[maxLayer][2];

	memcpy(&s[0], &k[1], 16);
	t[0] = k[17];

	for(int i = 1; i <= maxLayer; i++){
		memcpy(&sCW[i-1], &k[18 * i], 16);
		tCW[i-1][0] = k[18 * i + 16];
		tCW[i-1][1] = k[18 * i + 17];
	}

	uint128_t sL, sR;
	int tL, tR;
	for(int i = 1; i <= maxLayer; i++){
		dpfPRG(ctx, s[i - 1], &sL, &sR, &tL, &tR);

		if(t[i-1] == 1){
			sL = sL ^ sCW[i-1];
			sR = sR ^ sCW[i-1];
			tL = tL ^ tCW[i-1][0];
			tR = tR ^ tCW[i-1][1];
		}

		int xbit = getbit(x, n, i);
		if(xbit == 0){
			s[i] = sL;
			t[i] = tL;
		}else{
			s[i] = sR;
			t[i] = tR;
		}
	}

    //get the data share out
    int len = 0;
    uint8_t *zeros = (uint8_t*) malloc(dataSize+16);
    memset(zeros, 0, dataSize+16);
    //use a counter mode encryption of 0 with each seed as key to get prg output
    //printf("here\n");

    EVP_CIPHER_CTX *seedCtx;
    if(!(seedCtx = EVP_CIPHER_CTX_new()))
        printf("errors occured in creating context\n");
    if(1 != EVP_EncryptInit_ex(seedCtx, EVP_aes_128_ctr(), NULL, (uint8_t*)&s[maxLayer], NULL))
        printf("errors occured in init of dpf eval\n");
    if(1 != EVP_EncryptUpdate(seedCtx, dataShare, &len, zeros, ((dataSize-1)|15)+1))
        printf("errors occured in encrypt\n");
    
    for(int i = 0; i < dataSize; i++){
        if(t[maxLayer] == 1){
            //xor in correction word
            dataShare[i] = dataShare[i] ^ k[18*n+18+i];

            //printf("xoring stuff in at index %d\n", i);
        }
                //printf("%x\n", (*dataShare)[i]);

    }

    free(zeros);
    EVP_CIPHER_CTX_free(seedCtx);

    //print_block(s[maxLayer]);
    //printf("%x\n", t[maxLayer]);

    //use the last seed for dpf checking
	return s[maxLayer];
}

/* Need to allow specifying start and end for dataShare */
void evalAllDPF(EVP_CIPHER_CTX *ctx, int domainSize, unsigned char* k, int dataSize, uint8_t** dataShare){

    //dataShare is of size dataSize
    int numLeaves = pow(2, domainSize);
    int n = domainSize;
	int maxLayer = n;

    int currLevel = 0;
    int levelIndex = 0;
    int numIndexesInLevel = 2;

    int treeSize = 2 * numLeaves - 1;
    //int treeSize = 2 * (endIndex - startIndex) - 1;

	uint128_t s[treeSize];
	int t[treeSize];
	uint128_t sCW[maxLayer];
	int tCW[maxLayer][2];

	memcpy(&s[0], &k[1], 16);
	t[0] = k[17];

	for(int i = 1; i <= maxLayer; i++){
		memcpy(&sCW[i-1], &k[18 * i], 16);
		tCW[i-1][0] = k[18 * i + 16];
		tCW[i-1][1] = k[18 * i + 17];
	}

	uint128_t sL, sR;
	int tL, tR;
	for(int i = 1; i < treeSize; i+=2){
        int parentIndex = 0;
        if (i > 1) {
            parentIndex = i - levelIndex - ((numIndexesInLevel - levelIndex) / 2);
        }
        dpfPRG(ctx, s[parentIndex], &sL, &sR, &tL, &tR);

		if(t[parentIndex] == 1){
			sL = sL ^ sCW[currLevel];
			sR = sR ^ sCW[currLevel];
			tL = tL ^ tCW[currLevel][0];
			tR = tR ^ tCW[currLevel][1];
		}

        int lIndex =  i;
        //int lIndex =  i + (numIndexesInLevel - levelIndex) + (levelIndex * 2);
        int rIndex =  i + 1;
        //int rIndex =  i + (numIndexesInLevel - levelIndex) + (levelIndex * 2) + 1;
        s[lIndex] = sL;
        t[lIndex] = tL;
        s[rIndex] = sR;
        t[rIndex] = tR;

	    
        levelIndex += 2;
        if (levelIndex == numIndexesInLevel) {
            currLevel++;
            numIndexesInLevel *= 2;
            levelIndex = 0;
        }
    }

    //get the data share out
    uint8_t *zeros = (uint8_t*) malloc(dataSize+16);
    memset(zeros, 0, dataSize+16);
    //use a counter mode encryption of 0 with each seed as key to get prg output
    //printf("here\n");

    EVP_CIPHER_CTX *seedCtx;
    if(!(seedCtx = EVP_CIPHER_CTX_new()))
        printf("errors occured in creating context\n");

    for (int i = 0; i < numLeaves; i++) {
        int len = 0;
        int index = treeSize - numLeaves + i;
        
        if(1 != EVP_EncryptInit_ex(seedCtx, EVP_aes_128_ctr(), NULL, (uint8_t*)&s[index], NULL))
            printf("errors occured in init of dpf eval\n");
        if(1 != EVP_EncryptUpdate(seedCtx, dataShare[i], &len, zeros, ((dataSize-1)|15)+1))
            printf("errors occured in encrypt\n");

        for(int j = 0; j < dataSize; j++){
            if(t[index] == 1){
                //xor in correction word
                dataShare[i][j] = dataShare[i][j] ^ k[18*n+18+j];

                //printf("xoring stuff in at index %d\n", i);
            }
                //printf("%x\n", (*dataShare)[i]);
        }
    }

    free(zeros);
    EVP_CIPHER_CTX_free(seedCtx);

    //print_block(s[maxLayer]);
    //printf("%x\n", t[maxLayer]);

}
