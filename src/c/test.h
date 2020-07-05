#ifndef _TEST
#define _TEST

#include "params.h"

int makeRandKeywords(char **keywords);
int runUpdates(client *c, server *s1, server *s2, int index, char **keywords, int keywordsLen);
int runSearch (client *c, server *s1, server *s2, char *keyword, uint8_t docsPresent[NUM_DOCS_BYTES]);
void printTables(server *s1, server *s2);
int parseArgs(int argc, const char *argv[]);

int runUpdates_malicious(client *c, server *s1, server *s2, int index, char **keywords, int keywordsLen);
int runSearch_malicious(client *c, server *s1, server *s2, char *keyword, uint8_t docsPresent[NUM_DOCS_BYTES]);

#endif
