#include ../../CONFIG.mine

CC=gcc
CFLAGS= -Wall
#CFLAGS= -Wall -I/usr/local/opt/openssl@1.1/include
#CFLAGS= -Wall -O2 -I/usr/local/opt/openssl@1.1/include
LDFLAGS= -lcrypto -lssl -lm -lpthread libstemmer.o 
#LDFLAGS= -L/usr/local/opt/openssl@1.1/lib -lssl -lcrypto -lssl -lm 
SOURCES= dpf.c client.c server.c params.c test.c correctness_tests.c tokenize.o
OBJECTS=$(SOURCES:.c=.o)
#SOURCES_DEPS= dpf.c client.c server.c params.c
#OBJECTS_DEPS=$(SOURCES_DEPS:.c=.o)
EXECUTABLES=correctness_tests
LIB=libfss.a

all: $(SOURCES) $(LIB) $(EXECUTABLES) 

$(EXECUTABLES): dpf.o client.o server.o params.o test.o tokenize.o
	$(CC) -o $@ $^ $@.o $(LDFLAGS)

$(LIB): dpf.o client.o server.o params.o test.o correctness_tests.o tokenize.o
	ar rcs $@ $^

.c.o:
	$(CC) $(CFLAGS) $< -c -o $@

.PHONY: clean
clean:
	mv libstemmer.o libstemmer.old
	rm -f *.o
	mv libstemmer.old libstemmer.o
#	rm -f *.o $(LIB) $(OBJECTS) $(EXECUTABLE)
