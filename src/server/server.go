package main

/*
#include "../c/dpf.c"
#include "../c/dpf.h"
#include "../c/params.h"
#include "../c/params.c"
#include "../c/server.c"
#include "../c/server.h"
#include "../c/client.c"
#include "../c/client.h"
#include "../c/tokenize.c"
#include "../c/tokenize.h"
*/
import "C"
import (
    "log"
    "net"
    "common"
    "os"
    "fmt"
    "encoding/json"
    "encoding/hex"
    "bufio"
    "github.com/hashicorp/go-msgpack/codec"
    "unsafe"
    "flag"
    "sync"
    "time"
    "strconv"
    "io"
    "bytes"
)

const kNumThreads = 16

var sOld *C.server
var sNew *C.server

var config common.ServerConfig

var oldVersionNum int
var newVersionNum int
var incomingVersionNum int

var incomingUpdates map[int]common.Update
var isMalicious bool

/* Read in config file. */
func setupConfig(filename string) (common.ServerConfig, error) {
    config := common.ServerConfig{}
    file, err := os.Open(filename)
    if err != nil {
        return config, err
    }
    defer file.Close()
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&config)
    if err != nil {
        return config, err
    }
    return config, nil
}

/* Dispatch message to different handlers. */
func handleConnection(conn net.Conn) {
    defer conn.Close()
    r := bufio.NewReader(conn)
    w := bufio.NewWriter(conn)
    dec := codec.NewDecoder(r, &codec.MsgpackHandle{})
    enc := codec.NewEncoder(w, &codec.MsgpackHandle{})
    for {
        rpcType, err := r.ReadByte()
        if err != nil {
            if err.Error() == "EOF" {
                break
            } else {
                log.Fatalln(err)
            }
        }

        switch rpcType {

        case common.SEARCH_REQUEST_MALICIOUS:
            var req common.SearchRequest_malicious
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            start := time.Now()
            resp, respErr := searchKeyword_malicious(req)
            elapsed := time.Since(start)
            resp.ServerLatency = elapsed
            logLatency(elapsed, "malicious")
            if err := enc.Encode(&respErr); err != nil {
                log.Fatalln(err)
            }
            if err := enc.Encode(&resp); err != nil {
                log.Fatalln(err)
            }
            if err := w.Flush(); err != nil {
                log.Fatalln(err)
            }


        case common.SEARCH_REQUEST_SEMIHONEST:
            var req common.SearchRequest_semihonest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            start := time.Now()
            resp, respErr := searchKeyword_semihonest(req)
            elapsed := time.Since(start)
            resp.ServerLatency = elapsed
            logLatency(elapsed, "semihonest")
            if err := enc.Encode(&respErr); err != nil {
                log.Fatalln(err)
            }
            if err := enc.Encode(&resp); err != nil {
                log.Fatalln(err)
            }
            if err := w.Flush(); err != nil {
                log.Fatalln(err)
            }

        case common.UPDATE_REQUEST_MALICIOUS:
            var req common.UpdateRequest_malicious
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            updateDoc_malicious(req)

        case common.UPDATE_REQUEST_SEMIHONEST:
            var req common.UpdateRequest_semihonest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            updateDoc_semihonest(req)

        case common.SETUP_REQUEST:
            var req common.SetupRequest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            resp, respErr := fastSetup(req)
            if err := enc.Encode(&respErr); err != nil {
                log.Fatalln(err)
            }
            if err := enc.Encode(&resp); err != nil {
                log.Fatalln(err)
            }
            if err := w.Flush(); err != nil {
                log.Fatalln(err)
            }

        case common.BATCH_START_REQUEST:
            var req common.BatchStartRequest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            resp, respErr := startUpdateBatch(req)
            if err := enc.Encode(&respErr); err != nil {
                log.Fatalln(err)
            }
            if err := enc.Encode(&resp); err != nil {
                log.Fatalln(err)
            }
            if err := w.Flush(); err != nil {
                log.Fatalln(err)
            }

        case common.BATCH_FINISH_REQUEST:
            var req common.BatchFinishRequest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            resp, respErr := finishUpdateBatch(req)
            if err := enc.Encode(&respErr); err != nil {
                log.Fatalln(err)
            }
            if err := enc.Encode(&resp); err != nil {
                log.Fatalln(err)
            }
            if err := w.Flush(); err != nil {
                log.Fatalln(err)
            }

        default:
            log.Fatalln(fmt.Errorf("Unknown request type %d", rpcType))
        }
    }

}

/* Log latency to file in separate thread. */
func logLatency (latency time.Duration, tag string) {
    go func(latency time.Duration, tag string) {
        file, _ := os.Create(config.OutDir + "/" + strconv.Itoa(int(C.NUM_DOCS)) + "_docs_" + strconv.Itoa(int(C.BLOOM_FILTER_SZ)) + "bf_latency_" + tag)
        defer file.Close()
        io.WriteString(file, latency.String())
    }(latency, tag)
}

/* Process search request (malicious adversaries). */
func searchKeyword_malicious(req common.SearchRequest_malicious) (common.SearchResponse_malicious, error) {
    var s *C.server = sNew
    if (req.Version == newVersionNum) {
        s = sNew
        log.Println("new version num")
    } else if (req.Version == oldVersionNum) {
        s =  sOld
        log.Println("old version num")
    }  else {
        log.Println("Unknown version number: ", req.Version)
    }

    cKeys := (***C.uchar)(C.malloc(C.size_t(kNumThreads) * C.size_t(unsafe.Sizeof(&req))))
    cResultsTemp := (***C.uint8_t)(C.malloc(C.size_t(kNumThreads) * C.size_t(unsafe.Sizeof(&req))))
    cResultsFinal := (**C.uint8_t)(C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&req))))
    cKeysIndexable := (*[1<<30 - 1]**C.uchar)(unsafe.Pointer(cKeys))
    cResultsTempIndexable := (*[1<<30 - 1]**C.uint8_t)(unsafe.Pointer(cResultsTemp))
    cResultsFinalIndexable := (*[1<<30 - 1]*C.uint8_t)(unsafe.Pointer(cResultsFinal))

    for i := 0; i < kNumThreads; i++ {
        cKeysIndexable[i] = (**C.uchar)(C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&req))))
        cResultsTempIndexable[i] = (**C.uint8_t)(C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&req))))
        cKeysIndexableInner := (*[1<<30 - 1]*C.uchar)(unsafe.Pointer(cKeysIndexable[i]))
        cResultsTempIndexableInner := (*[1<<30 - 1]*C.uint8_t)(unsafe.Pointer(cResultsTempIndexable[i]))
        for j := 0; j < int(C.BLOOM_FILTER_K); j++ {
            cKeysIndexableInner[j] = (*C.uchar)(C.CBytes(req.Keys[j]))
            cResultsTempIndexableInner[j] = (*C.uint8_t)(C.malloc((C.size_t(C.MALICIOUS_DPF_LEN))))
        }
    }

    for i := 0; i < int(C.BLOOM_FILTER_K); i++ {
        cResultsFinalIndexable[i] = (*C.uint8_t)(C.malloc((C.size_t(C.MALICIOUS_DPF_LEN))))
    }

    slice := int(C.BLOOM_FILTER_SZ) / kNumThreads
    var wg sync.WaitGroup
    wg.Add(kNumThreads)

    for i := 0; i < kNumThreads; i++ {
        go func(index int) {
            defer wg.Done()
            C.runQuery_malicious((*C.server)(s),
                                 (**C.uchar)(cKeysIndexable[index]),
                                 (**C.uint8_t)(cResultsTempIndexable[index]),
                                 C.int(index),
                                 C.int(index * slice),
                                 C.int((index + 1) * slice))
        }(i)
    }

    wg.Wait()

    C.assemblePerThreadResults((*C.server)(s),
                                (***C.uint8_t)(cResultsTemp),
                                (C.int)(kNumThreads),
                                (**C.uint8_t)(cResultsFinal))

    results := make([][]byte, C.int(C.BLOOM_FILTER_K))

    for i := 0; i < int(C.BLOOM_FILTER_K); i++ {
        results[i] = C.GoBytes(unsafe.Pointer(cResultsFinalIndexable[i]), C.int(C.MALICIOUS_DPF_LEN))
    }

    resp := common.SearchResponse_malicious{
        Results: results,
    }

    return resp, nil
}

/* Process search request (semihonest adversaries). */
func searchKeyword_semihonest(req common.SearchRequest_semihonest) (common.SearchResponse_semihonest, error) {
    var s *C.server = sNew
    if (req.Version == newVersionNum) {
        s = sNew
    } else if (req.Version == oldVersionNum) {
        s =  sOld
    }  else {
        log.Println("Unknown version number: ", req.Version)
    }

    cKeys := (***C.uchar)(C.malloc(C.size_t(kNumThreads) * C.size_t(unsafe.Sizeof(&req))))
    cResultsTemp := (***C.uint8_t)(C.malloc(C.size_t(kNumThreads) * C.size_t(unsafe.Sizeof(&req))))
    cResultsFinal := (**C.uint8_t)(C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&req))))
    cKeysIndexable := (*[1<<30 - 1]**C.uchar)(unsafe.Pointer(cKeys))
    cResultsTempIndexable := (*[1<<30 - 1]**C.uint8_t)(unsafe.Pointer(cResultsTemp))
    cResultsFinalIndexable := (*[1<<30 - 1]*C.uint8_t)(unsafe.Pointer(cResultsFinal))

    for i := 0; i < kNumThreads; i++ {
        cKeysIndexable[i] = (**C.uchar)(C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&req))))
        cResultsTempIndexable[i] = (**C.uint8_t)(C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&req))))
        cKeysIndexableInner := (*[1<<30 - 1]*C.uchar)(unsafe.Pointer(cKeysIndexable[i]))
        cResultsTempIndexableInner := (*[1<<30 - 1]*C.uint8_t)(unsafe.Pointer(cResultsTempIndexable[i]))
        for j := 0; j < int(C.BLOOM_FILTER_K); j++ {
            cKeysIndexableInner[j] = (*C.uchar)(C.CBytes(req.Keys[j]))
            cResultsTempIndexableInner[j] = (*C.uint8_t)(C.malloc((C.size_t(C.NUM_DOCS_BYTES))))
        }
    }

    for i := 0; i < int(C.BLOOM_FILTER_K); i++ {
        cResultsFinalIndexable[i] = (*C.uint8_t)(C.malloc((C.size_t(C.NUM_DOCS_BYTES))))
    }

    slice := int(C.BLOOM_FILTER_SZ) / kNumThreads
    var wg sync.WaitGroup
    wg.Add(kNumThreads)

    for i := 0; i < kNumThreads; i++ {
        go func(index int) {
            defer wg.Done()
            C.runQuery((*C.server)(s),
                        (**C.uchar)(cKeysIndexable[index]),
                        (**C.uint8_t)(cResultsTempIndexable[index]),
                        C.int(index),
                        C.int(index * slice),
                        C.int((index + 1) * slice))
        }(i)
    }

    wg.Wait()

    C.assemblePerThreadResults((*C.server)(s),
                                (***C.uint8_t)(cResultsTemp),
                                (C.int)(kNumThreads),
                                (**C.uint8_t)(cResultsFinal))

    results := make([][]byte, C.int(C.BLOOM_FILTER_K))

    for i := 0; i < int(C.BLOOM_FILTER_K); i++ {
        results[i] = C.GoBytes(unsafe.Pointer(cResultsFinalIndexable[i]), C.int(C.NUM_DOCS_BYTES))
    }

    resp := common.SearchResponse_semihonest{
        Results: results,
    }

    return resp, nil
}

/* Only called directly if not using the master; for updating documents (malicious adversaries). */
func updateDoc_malicious(req common.UpdateRequest_malicious) (common.UpdateResponse_malicious, error) {
    var s *C.server = sNew

    joinedMacs := bytes.Join(req.MACs, nil)
    cMacs := (*C.uint128_t)(C.CBytes(joinedMacs))

    C.setRow_malicious((*C.server)(s),
                       (C.int)(req.DocID),
                       (*C.uint8_t)(C.CBytes(req.BF)),
                       (*C.uint128_t)(cMacs));


    return common.UpdateResponse_malicious{Test: "success"}, nil

}

/* Only called directly if not using the master; for updating documents (semihonest adversaries). */
func updateDoc_semihonest(req common.UpdateRequest_semihonest) (common.UpdateResponse_semihonest, error) {
    var s *C.server = sNew

    C.setRow((*C.server)(s),
            (C.int)(req.DocID),
            (*C.uint8_t)(C.CBytes(req.BF)));


    return common.UpdateResponse_semihonest{Test: "success"}, nil
}

/* Run fast setup to initialize state for benchmarking and tests. */
func fastSetup(req common.SetupRequest) (common.SetupResponse, error) {
    c := &C.client{}
    maskKey := make([]byte, hex.DecodedLen(len(config.ClientMaskKey)))
    macKey := make([]byte, hex.DecodedLen(len(config.ClientMacKey)))
    hex.Decode(maskKey, []byte(config.ClientMaskKey))
    hex.Decode(macKey, []byte(config.ClientMacKey))
    C.initializeClient((*C.client)(c), C.int(kNumThreads), (*C.uint8_t)(C.CBytes(maskKey)), (*C.uint8_t)(C.CBytes(macKey)))

    if (req.BenchmarkDir == "") {
        keywords := []string{"hello", "world"}
        cKeywords := C.malloc(C.size_t(len(keywords)) * C.size_t (common.MAX_KEYWORD_SIZE))
        cKeywordsIndexable := (*[1<<30 - 1]*C.char)(cKeywords)
        for i, keyword := range keywords {
            cKeywordsIndexable[i] = C.CString(keyword)
        }
        buf := C.malloc(C.size_t(C.BLOOM_FILTER_BYTES))
        macLen := C.size_t(C.BLOOM_FILTER_SZ) * C.MAC_BYTES
        cMacs := C.malloc(macLen)
        for i := 0; i < int(C.MAX_DOCS); i++ {
            C.generateEncryptedBloomFilter((*C.client)(c),
                                            (*C.uint8_t)(buf),
                                            (**C.char)(cKeywords),
                                            (C.size_t)(len(keywords)),
                                            C.int(i),
                                            nil)

            C.generateMACsForBloomFilter_malicious((*C.client)(c),
                                           (*C.uint8_t)(buf),
                                           C.int(i),
                                           (*C.uint128_t)(cMacs))

            C.setRow_malicious((*C.server)(sOld),
                                (C.int)(i),
                                (*C.uint8_t)(buf),
                                (*C.uint128_t)(cMacs));

        }
    }
    C.copyServer(sNew, sOld)
    log.Println("Finished initialization")

    versions := make([]uint32, C.int(C.MAX_DOCS))
    cVersionsIndexable := (*[1<<30 - 1]C.uint32_t)(unsafe.Pointer(c.versions))

    for i := 0; i < int(C.MAX_DOCS); i++ {
        versions[i] = uint32(cVersionsIndexable[i])
    }

    resp := common.SetupResponse{
        NumDocs: int(C.NUM_DOCS),
        Versions: versions,
    }

    C.freeClient((*C.client)(c))

    return resp, nil
}

/* Start processing incoming batch of updates from master. */
func startUpdateBatch(req common.BatchStartRequest) (common.BatchStartResponse, error) {
    isMalicious = req.Malicious
    incomingUpdates = req.Updates
    incomingVersionNum = req.VersionNum
    return common.BatchStartResponse{Commit: true}, nil
}

/* Finish processing incoming batch of updates from master. */
func finishUpdateBatch(req common.BatchFinishRequest) (common.BatchFinishResponse, error) {
    C.copyServer((*C.server)(sOld), (*C.server)(sNew))
    oldVersionNum =  newVersionNum
    /* Process all incoming updates. */
    if (isMalicious) {
        for doc, update := range incomingUpdates {
            joinedMacs := bytes.Join(update.MACs, nil)
            if len(joinedMacs) == 0 {
                continue
            }
            cMacs := (*C.uint128_t)(C.CBytes(joinedMacs))

            C.setRow_malicious((*C.server)(sNew),
                                (C.int)(doc),
                                (*C.uint8_t)(C.CBytes(update.BF)),
                                (*C.uint128_t)(cMacs))
        }
    } else {
        for doc, update := range incomingUpdates {
             C.setRow((*C.server)(sNew),
                      (C.int)(doc),
                      (*C.uint8_t)(C.CBytes(update.BF)))
        }
    }
    newVersionNum = incomingVersionNum
    return common.BatchFinishResponse{}, nil
}


func main() {
    /* Set up config */
    filename := flag.String("config", "src/config/server1.config", "server config file")
    maxDocs := flag.Int("max_docs", 256, "max number of documents")
    bloomFilterSz := flag.Int("bf_sz", 128, "bloom filter size in bits")
    flag.Parse()

    var err error
    config, err = setupConfig(*filename)
    if err != nil {
        log.Fatalln("Error retrieving config file: ", err)
    }

    /* Initialize server */
    log.Println("Starting initialization...")
    C.setSystemParams(C.int(*bloomFilterSz), C.int(*maxDocs));
    sOld = &C.server{}
    sNew = &C.server{}
    C.initializeServer((*C.server)(sOld), C.int(kNumThreads))
    C.initializeServer((*C.server)(sNew), C.int(kNumThreads))
    oldVersionNum = 0
    newVersionNum = 0

    /* Start listening */
    log.Println("Listening...")
    err = common.ListenLoop(config.Port, config.CertFile, config.KeyFile, handleConnection)
    if err != nil {
        log.Fatalln(err)
    }
}
