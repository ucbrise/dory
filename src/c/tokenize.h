#ifndef _TOKENIZE
#define _TOKENIZE

#include <stdio.h>

#define MAX_NUM_KEYWORDS 1000

void initStemmer();
char *stemWord (char *keywordIn);
int tokenizeDoc(char **keywords, char **lines, int numLines);
int tokenizeFile(char **keywords, char *filename);

#endif
