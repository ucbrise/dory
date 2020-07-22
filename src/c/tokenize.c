#include "stdio.h"
#include <stdbool.h>
#include "libstemmer.h"
#include "tokenize.h"

struct sb_stemmer *stemmer;

void initStemmer() {
    stemmer = sb_stemmer_new("english", "UTF_8");
}

char *stemWord (char *keywordIn) {
    char *tmp = malloc(strlen(keywordIn));
    strcpy(tmp, keywordIn);
    for (int i = 0; i < strlen(keywordIn); i++) {
        if (isupper(keywordIn[i])) tmp[i] = tolower(keywordIn[i]);
    }
    const sb_symbol *stemmed = sb_stemmer_stem(stemmer, tmp, strlen(tmp));
    char *keywordOut = malloc(strlen((const char *)stemmed));
    strcpy(keywordOut, (const char *)stemmed);
    return keywordOut;
}

/* Tokenize an array of lines into keywords (maximum number of keywords). */
int tokenizeDoc(char **keywords, char **lines, int numLines) {
    int ctr = 0;
    const char *delims = ".;:!?/\\,#@$&)(\" =";
    for (int i = 0; i < numLines; i++) {
        char *token = strtok(lines[i], delims);
        while (token != NULL) {
            if (strlen(token) <= 3 || strlen(token) >= 20) break;
            bool shouldBreak = false;
            for (int j = 0; j < strlen(token); j++) {
                if (!isalpha(token[j])) {
                    shouldBreak = true;
                    break;
                }
                if (isupper(token[j])) token[j] = tolower(token[j]);
            }
            const sb_symbol *stemmed =  sb_stemmer_stem(stemmer, token, strlen(token));
            keywords[ctr] = malloc(strlen((const char *)stemmed));
            strcpy(keywords[ctr], (const char *)stemmed);
            token = strtok(NULL, delims);
            ctr++;
            if (ctr >= MAX_NUM_KEYWORDS) return ctr;
        }
    }
    return ctr;
}

/* Tokenize a file into keywords (maximum number of keywords). */
int tokenizeFile(char **keywords, char *filename) {
    FILE *fp = fopen(filename, "r");
    if (fp == NULL) {
        return 0;
    }
    int ctr = 0;
    char *line = NULL;
    size_t len = 0; 
    ssize_t read;
    const char *delims = ".;:!?/\\,#@$&)(\" =";
    while ((read = getline(&line, &len, fp)) != -1) {
        char *token = strtok(line, delims);
        while (token != NULL) {
            if (strlen(token) <= 3 || strlen(token) >= 20) break;
            bool shouldBreak = false;
            for (int j = 0; j < strlen(token); j++) {
                if (!isalpha(token[j])) {
                    shouldBreak = true;
                    break;
                }
                if (isupper(token[j])) token[j] = tolower(token[j]);
            }
            const sb_symbol *stemmed =  sb_stemmer_stem(stemmer, token, strlen(token));
            keywords[ctr] = malloc(strlen((const char *)stemmed));
            strcpy(keywords[ctr], (const char *)stemmed);
            token = strtok(NULL, delims);
            ctr++;
            if (ctr >= MAX_NUM_KEYWORDS) {
                if (line) free(line);
                fclose(fp);
                return ctr;
            }
        }
    }
    fclose(fp);
    if (line) {
        free(line);
    }
    return ctr;
}
