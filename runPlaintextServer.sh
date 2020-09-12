CGO_CFLAGS="-I./libsolv-sys/src -D LIBSOLV_INTERNAL -w" CGO_LDFLAGS="-lssl -lpthread -lcrypto -lm "$PWD"/src/c/libstemmer.o" go run src/plaintext_server/plaintext_server.go
