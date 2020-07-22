# DORY: An Encrypted Search System with Distributed Trust

DORY is an encrypted search system that splits trust between multiple servers in order to efficiently hide access patterns from a malicious attacker who controls all but one of the servers. This implementation does NOT include the underlying end-to-end encrypted filesystem.

**WARNING**: This is an academic proof-of-concept prototype and has not received careful code review. This implementation is NOT ready for production use.

This prototype is released under the Apache v2 license (see [License](#license)).

## Setup
1. Install OpenSSL, tested up to 2.6.5.
2. Run `go get github.com/hashicorp/go-msgpack/codec`.
3. Download and build `libstemmer` (http://snowball.tartarus.org/download.html), tested up to version 2.0.0.
4. Move the output `libstemmer.o` to `src/c/`.

## Configuration
Update the config files in `src/config/` to reflect the IP addresses and ports for the master, servers, and client. Default TLS keys and certificates are included for testing, although these should be freshly generated for security. To use config files with different names, update `runMaster.sh`, `runServer.sh`, and/or `runClient.sh`.

## Usage
Start the master by running `runMaster.sh`, the servers by running `runServer.sh` and the client by running `runClient.sh`. Each script has a number of flags that can be set; run the scripts with `-h` to see all the flags.

For example, to start test DORY on a single machine (with two servers), use the default config files and run:
```
./runMaster.sh
./runServer.sh -s 1
./runServer.sh -s 2
./runClient.sh
```

Without any flags set, the client will load all the documents in `sample_docs` into the search index and then provide prompts for the user to enter keywords to search for.

Make sure to always set the Bloom filter size and the max number of documents the same across the master, servers, and clients. The only exception is when running with cluster sizes greater than 1; in this case, every entity should use the same Bloom filter size, the master and client should use the correct maximum number of documents, and the servers should use the maximum  number of documents divided by the number of clusters. To run with multiple clusters, you will need a number of servers equal to 2 times the number of clusters.

## Tests
To test the low-level crypto, run `make` in `src/c` and run `correctness_tests`. To test the end-to-end system, run the client with the correctness test flag set to true (`-c`).

## Benchmarks
Use the scripts in `bench/` to run latency and throughput benchmarks on the system. Update the scripts with the IP addresses and ports of the different entities before running.

## Acknowledgements

We build on Saba Eskandarian's DPF implementation in https://github.com/SabaEskandarian/Express.

