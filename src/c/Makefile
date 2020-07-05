include ../../CONFIG.mine

CC=gcc
CFLAGS= -Wall -I$(OSSL)/include
#CFLAGS= -Wall -I/usr/local/opt/openssl@1.1/include
#CFLAGS= -Wall -O2 -I/usr/local/opt/openssl@1.1/include
LDFLAGS= -L$(OSSL)/lib  -lssl -lcrypto -lssl -lm libstemmer.o 
#LDFLAGS= -L/usr/local/opt/openssl@1.1/lib -lssl -lcrypto -lssl -lm 
SOURCES= dpf.c client.c server.c params.c test.c bench.c correctness_tests.c aes_bench.c tokenize.o
OBJECTS=$(SOURCES:.c=.o)
#SOURCES_DEPS= dpf.c client.c server.c params.c
#OBJECTS_DEPS=$(SOURCES_DEPS:.c=.o)
EXECUTABLES=bench correctness_tests aes_bench
LIB=libfss.a

all: $(SOURCES) $(LIB) $(EXECUTABLES) 

$(EXECUTABLES): dpf.o client.o server.o params.o test.o tokenize.o
	$(CC) -o $@ $^ $@.o $(LDFLAGS)

$(LIB): dpf.o client.o server.o params.o test.o bench.o correctness_tests.o aes_bench.o tokenize.o
	ar rcs $@ $^

.c.o:
	$(CC) $(CFLAGS) $< -c -o $@

.PHONY: clean
clean:
	rm -f *.o
#	rm -f *.o $(LIB) $(OBJECTS) $(EXECUTABLE)
