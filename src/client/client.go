package client

/*
#include "../c/dpf.c"
#include "../c/dpf.h"
#include "../c/params.h"
#include "../c/params.c"
#include "../c/client.c"
#include "../c/client.h"
#include "../c/tokenize.h"
#include "../c/tokenize.c"
*/
import "C"
import (
    "common"
    "encoding/json"
    "encoding/hex"
    "log"
    "unsafe"
    "os"
    "io"
    "time"
    "sync"
    "strconv"
)

const kNumThreads = 1

var c *C.client
var clientLock sync.Mutex
var config common.ClientConfig

var versionNum int

/* Read in config file. */
func setupConfig(filename string) (common.ClientConfig, error) {
    config := common.ClientConfig{}
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

/* Log latency to file in separate thread. */
func logLatency (latency time.Duration, tag string) {
    go func(latency time.Duration, tag string) {
        file, _ := os.Create(config.OutDir + "/" + strconv.Itoa(int(C.NUM_DOCS)) + "_docs_" + strconv.Itoa(int(C.BLOOM_FILTER_SZ)) + "bf_latency_" + tag)
        defer file.Close()
        io.WriteString(file, latency.String())
    }(latency, tag)
}

/* Run update (malicious adversaries). */
func UpdateDoc_malicious(conn *common.Conn, keywords []string, docID int, useMaster bool) error {
    /* Convert []string to **char */
    cKeywords := C.malloc(C.size_t(len(keywords)) * C.size_t(unsafe.Sizeof(&docID)))
    cKeywordsIndexable := (*[1<<30 - 1]*C.char)(cKeywords)
    for i, keyword := range keywords {
        cKeywordsIndexable[i] = C.CString(keyword)
    }
    cVersionNum := C.uint32_t(0)

    /* Generate encrypted bloom filter */
    buf := C.malloc(C.size_t(C.BLOOM_FILTER_BYTES))
    clientLock.Lock()
    C.generateEncryptedBloomFilter((*C.client)(c),
                                   (*C.uint8_t)(buf),
                                   (**C.char)(cKeywords),
                                   (C.size_t)(len(keywords)),
                                   C.int(docID),
                                   (*C.uint32_t)(&cVersionNum))

    /* Generate MACs */
    macLen := C.size_t(C.BLOOM_FILTER_SZ) * C.MAC_BYTES
    cMacs := C.malloc(macLen)
    cMacsIndexable := (*[1<<30 - 1]C.uint128_t)(cMacs)
    C.generateMACsForBloomFilter_malicious((*C.client)(c),
                                           (*C.uint8_t)(buf),
                                           C.int(docID),
                                           (*C.uint128_t)(cMacs))
    clientLock.Unlock()
    macs := make([][]byte, C.BLOOM_FILTER_SZ)
    for i := 0; i < int(C.BLOOM_FILTER_SZ); i++ {
        macs[i] = C.GoBytes(unsafe.Pointer(&cMacsIndexable[i]),
                            C.MAC_BYTES)
    }

    /* Make request for servers */
    req := &common.UpdateRequest_malicious{
        DocID: docID,
        Version: uint32(cVersionNum),
        BF: C.GoBytes(buf, C.BLOOM_FILTER_BYTES),
        MACs: macs,
    }

    if (useMaster) {
        common.SendMessageWithConnectionNoResp(
            conn,
            common.UPDATE_REQUEST_MALICIOUS,
            req,
        )
    } else {
        var wg sync.WaitGroup
        wg.Add(2)

        /* Send to server 1 */
        go func() {
            defer wg.Done()
            common.SendMessageNoResp(
                config.Addr[0] + config.Port[0],
                common.UPDATE_REQUEST_MALICIOUS,
                req,
            )
        }()

        /* Send request to server 2 */
        go func() {
            defer wg.Done()
            common.SendMessageNoResp(
                config.Addr[1] + config.Port[1],
                common.UPDATE_REQUEST_MALICIOUS,
                req,
            )
        }()
        wg.Wait()
    }

    /* Free allocated memory */
    C.free(cKeywords)
    C.free(buf)
    C.free(cMacs)

    return nil
}

/* Send dummy update (only used for throughput measurements for malicious adversaries). */
func DummyUpdateDoc_malicious(conn *common.Conn, keywords []string, docID int, useMaster bool) error {

    macs := make([][]byte, C.BLOOM_FILTER_SZ)
    for i := 0; i < int(C.BLOOM_FILTER_SZ); i++ {
        macs[i] = make([]byte, C.MAC_BYTES)
    }

    /* Make request for servers */
    req := &common.UpdateRequest_malicious{
        DocID: docID,
        Version: 0,
        BF: make([]byte, C.BLOOM_FILTER_BYTES),
        MACs: macs,
    }

    if (useMaster) {
        common.SendMessageWithConnectionNoResp(
            conn,
            common.UPDATE_REQUEST_MALICIOUS,
            req,
        )
    } else {
        var wg sync.WaitGroup
        wg.Add(2)

        /* Send to server 1 */
        go func() {
            defer wg.Done()
            common.SendMessageNoResp(
                config.Addr[0] + config.Port[0],
                common.UPDATE_REQUEST_MALICIOUS,
                req,
            )
        }()

        /* Send request to server 2 */
        go func() {
            defer wg.Done()
            common.SendMessageNoResp(
                config.Addr[1] + config.Port[1],
                common.UPDATE_REQUEST_MALICIOUS,
                req,
            )
        }()
        wg.Wait()
    }

    return nil
}

/* Generate update (semihonest adversaries). */
func UpdateDoc_semihonest(conn *common.Conn, keywords []string, docID int, useMaster bool) error {
    /* Convert []string to **char */
    cKeywords := C.malloc(C.size_t(len(keywords)) * C.size_t(common.MAX_KEYWORD_SIZE))
    cKeywordsIndexable := (*[1<<30 - 1]*C.char)(cKeywords)
    for i, keyword := range keywords {
        cKeywordsIndexable[i] = C.CString(keyword)
    }
    cVersionNum := C.uint32_t(0)

    /* Generate encrypted bloom filter */
    buf := C.malloc(C.size_t(C.BLOOM_FILTER_BYTES))
    clientLock.Lock()
    C.generateEncryptedBloomFilter((*C.client)(c),
                                   (*C.uint8_t)(buf),
                                   (**C.char)(cKeywords),
                                   (C.size_t)(len(keywords)),
                                   C.int(docID),
                                   (*C.uint32_t)(&cVersionNum))
    clientLock.Unlock()

    /* Make request for servers */
    req := &common.UpdateRequest_semihonest{
        DocID: docID,
        Version: uint32(cVersionNum),
        BF: C.GoBytes(buf, C.BLOOM_FILTER_BYTES),
    }

    if (useMaster) {
        common.SendMessageWithConnectionNoResp(
            conn,
            common.UPDATE_REQUEST_SEMIHONEST,
            req,
        )
    } else {
        var wg sync.WaitGroup
        wg.Add(2)

        /* Send to server 1 */
        go func() {
            defer wg.Done()
            common.SendMessageNoResp(
                config.Addr[0] + config.Port[0],
                common.UPDATE_REQUEST_SEMIHONEST,
                req,
            )
        }()

        /* Send request to server 2 */
        go func() {
            defer wg.Done()
            common.SendMessageNoResp(
                config.Addr[1] + config.Port[1],
                common.UPDATE_REQUEST_SEMIHONEST,
                req,
            )
        }()

        wg.Wait()
    }

    /* Free allocated memory */
    C.free(cKeywords)
    C.free(buf)
    return nil
}

/* Send dummy update (only used for throughput measurements for semihonest adversaries). */
func DummyUpdateDoc_semihonest(conn *common.Conn, keywords []string, docID int, useMaster bool) error {
    /* Make request for servers */
    req := &common.UpdateRequest_semihonest{
        DocID: docID,
        Version: 0,
        BF: make([]byte, C.BLOOM_FILTER_BYTES),
    }

    if (useMaster) {
        common.SendMessageWithConnectionNoResp(
            conn,
            common.UPDATE_REQUEST_SEMIHONEST,
            req,
        )
    } else {
        var wg sync.WaitGroup
        wg.Add(2)

        /* Send to server 1 */
        go func() {
            defer wg.Done()
            common.SendMessageNoResp(
                config.Addr[0] + config.Port[0],
                common.UPDATE_REQUEST_SEMIHONEST,
                req,
            )
        }()

        /* Send request to server 2 */
        go func() {
            defer wg.Done()
            common.SendMessageNoResp(
                config.Addr[1] + config.Port[1],
                common.UPDATE_REQUEST_SEMIHONEST,
                req,
            )
        }()

        wg.Wait()
    }

    /* Free allocated memory */
    return nil
}

func StemWord(keyword string) string {
    cKeyword := C.CString(keyword)
    return C.GoString(C.stemWord(cKeyword))
}

func GetKeywordsFromFile(filename string) []string {
    cFilename := C.CString(filename)
    cKeywords := (**C.char)(C.malloc(C.MAX_NUM_KEYWORDS * C.size_t(unsafe.Sizeof(&filename))))
    cKeywordsIndexable := (*[1<<30 - 1]*C.char)(unsafe.Pointer(cKeywords))
    numKeywords := C.tokenizeFile(cKeywords, cFilename)

    keywords := make([]string, numKeywords)
    for i := 0; i < int(numKeywords); i++ {
        keywords[i] = C.GoString(cKeywordsIndexable[i])
    }
    return keywords
}

/* Update document, tokenizing words from file (malicious adveraries). */
func UpdateDocFile_malicious(conn *common.Conn, filename string, docID int, useMaster bool) error {
    keywords := GetKeywordsFromFile(filename)
    return UpdateDoc_malicious(conn, keywords, docID, useMaster)
}

/* Update document, tokenizing words from file (semihonest adveraries). */
func UpdateDocFile_semihonest(conn *common.Conn, filename string, docID int, useMaster bool) error {
    keywords := GetKeywordsFromFile(filename)
    return UpdateDoc_semihonest(conn, keywords, docID, useMaster)
}

func SearchKeyword_malicious(conn *common.Conn, keyword string, useMaster bool) ([]byte, error, time.Duration, time.Duration, time.Duration, time.Duration, time.Duration) {
    t1 := time.Now()

    if (useMaster) {
        GetState(conn)
    }

    t2 := time.Now()
    
    // In progress
    cKeys1 := (**C.uchar)(C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&keyword))))
    cKeys2 := (**C.uchar)(C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&keyword))))
    cResults1 := C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&keyword)))
    cResults2 := C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&keyword)))
    cIndexes := C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof((C.uint32_t)(1))))
    cDocsPresent := C.malloc(C.size_t(C.NUM_DOCS_BYTES))
    cKeys1Indexable := (*[1<<30 - 1]*C.uchar)(unsafe.Pointer(cKeys1))
    cKeys2Indexable := (*[1<<30 - 1]*C.uchar)(unsafe.Pointer(cKeys2))
    cResults1Indexable := (*[1<<30 - 1]*C.uint8_t)(unsafe.Pointer(cResults1))
    cResults2Indexable := (*[1<<30 - 1]*C.uint8_t)(unsafe.Pointer(cResults2))
   
    clientLock.Lock()
    C.generateKeywordQuery_malicious((*C.client)(c),
                                     C.CString(keyword),
                                     (**C.uchar)(cKeys1),
                                     (**C.uchar)(cKeys2),
                                     (*C.uint32_t)(cIndexes))
    clientLock.Unlock()

    keys1 := make([][]byte, int(C.BLOOM_FILTER_K))
    keys2 := make([][]byte, int(C.BLOOM_FILTER_K))

    for i := 0; i < int(C.BLOOM_FILTER_K); i++ {
        keys1[i] = C.GoBytes(unsafe.Pointer(cKeys1Indexable[i]), C.int(C.getDPFKeyLen_malicious()))
        keys2[i] = C.GoBytes(unsafe.Pointer(cKeys2Indexable[i]), C.int(C.getDPFKeyLen_malicious()))
    }

    req1 := &common.SearchRequest_malicious{
        Keys: keys1,
        Version: versionNum,
    }
    req2 := &common.SearchRequest_malicious{
        Keys: keys2,
        Version: versionNum,
    }

    t3 := time.Now()

    var wg sync.WaitGroup
    wg.Add(2)

    var serverLatency time.Duration

    resp1 := &common.SearchResponse_malicious{}
    var respError1 error
    go func() {
        defer wg.Done()
        common.SendMessage(
            config.Addr[0] + config.Port[0],
            common.SEARCH_REQUEST_MALICIOUS,
            req1,
            resp1,
            &respError1,
        )
        if (resp1.ServerLatency > serverLatency) {
            serverLatency = resp1.ServerLatency
        }
    }()

    resp2 := &common.SearchResponse_malicious{}
    var respError2 error
    go func() {
        defer wg.Done()
        common.SendMessage(
            config.Addr[1] + config.Port[1],
            common.SEARCH_REQUEST_MALICIOUS,
            req2,
            resp2,
            &respError2,
        )
        if (resp2.ServerLatency > serverLatency) {
            serverLatency = resp2.ServerLatency
        }
    }()
  
    wg.Wait()

    t4 := time.Now()

    for i := 0; i < int(C.BLOOM_FILTER_K); i++ {
       cResults1Indexable[i] = (*C.uint8_t)(C.CBytes(resp1.Results[i]))
       cResults2Indexable[i] = (*C.uint8_t)(C.CBytes(resp2.Results[i]))
    }
  
    clientLock.Lock()
    start := time.Now()
    C.assembleQueryResponses_malicious((*C.client)(c),
                                       (**C.uint8_t)(cResults1),
                                       (**C.uint8_t)(cResults2),
                                       (*C.uint32_t)(cIndexes),
                                       (*C.uint8_t)(cDocsPresent))
    elapsed := time.Since(start)
    clientLock.Unlock()
    logLatency(elapsed, "reconstruct_malicious")


    docsPresent := C.GoBytes(cDocsPresent, C.int(C.NUM_DOCS_BYTES))

    t5 := time.Now()

    return docsPresent, nil, t2.Sub(t1), t3.Sub(t2), t4.Sub(t3), serverLatency, t5.Sub(t4)
}

/* Run search, possibly distributed across clusters (malicious adversaries). */
func SearchKeyword_malicious_parallel(conn *common.Conn, keyword string, useMaster bool, numClusters int) ([]byte, error) {

    if (useMaster) {
        GetState(conn)
    }


    var wg sync.WaitGroup
    wg.Add(2 * numClusters)
    cIndexes := make([]unsafe.Pointer, numClusters)

    resps := make([]*common.SearchResponse_malicious, numClusters * 2)

    for i := 0; i < numClusters; i++ {
        // In progress
        cKeys1 := (**C.uchar)(C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&keyword))))
        cKeys2 := (**C.uchar)(C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&keyword))))
        cIndexes[i] = C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof((C.uint32_t)(1))))
        cKeys1Indexable := (*[1<<30 - 1]*C.uchar)(unsafe.Pointer(cKeys1))
        cKeys2Indexable := (*[1<<30 - 1]*C.uchar)(unsafe.Pointer(cKeys2))
   
        clientLock.Lock()
        C.generateKeywordQuery_malicious((*C.client)(c),
                                     C.CString(keyword),
                                     (**C.uchar)(cKeys1),
                                     (**C.uchar)(cKeys2),
                                     (*C.uint32_t)(cIndexes[i]))
        clientLock.Unlock()

        keys1 := make([][]byte, int(C.BLOOM_FILTER_K))
        keys2 := make([][]byte, int(C.BLOOM_FILTER_K))

        for i := 0; i < int(C.BLOOM_FILTER_K); i++ {
            keys1[i] = C.GoBytes(unsafe.Pointer(cKeys1Indexable[i]), C.int(C.getDPFKeyLen_malicious()))
            keys2[i] = C.GoBytes(unsafe.Pointer(cKeys2Indexable[i]), C.int(C.getDPFKeyLen_malicious()))
        }

        req1 := &common.SearchRequest_malicious{
            Keys: keys1,
            Version: versionNum,
        }
        req2 := &common.SearchRequest_malicious{
            Keys: keys2,
            Version: versionNum,
        }

        resps[2 * i] = &common.SearchResponse_malicious{}
        var respError1 error
        go func(i int) {
            defer wg.Done()
            common.SendMessage(
                config.Addr[2 * i] + config.Port[2 * i],
                common.SEARCH_REQUEST_MALICIOUS,
                req1,
                resps[2 * i],
                &respError1,
            )
        }(i)

        resps[2 * i + 1] = &common.SearchResponse_malicious{}
        var respError2 error
        go func(i int) {
            defer wg.Done()
            common.SendMessage(
                config.Addr[2 * i + 1] + config.Port[2 * i + 1],
                common.SEARCH_REQUEST_MALICIOUS,
                req2,
                resps[2 * i + 1],
                &respError2,
            )
        }(i)
    }

    wg.Wait()

    var docsPresent []byte

    for i := 0; i < numClusters; i++ {

        cResults1 := C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&keyword)))
        cResults2 := C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&keyword)))
        cResults1Indexable := (*[1<<30 - 1]*C.uint8_t)(unsafe.Pointer(cResults1))
        cResults2Indexable := (*[1<<30 - 1]*C.uint8_t)(unsafe.Pointer(cResults2))
        cDocsPresent := C.malloc(C.size_t(C.NUM_DOCS_BYTES))

        for j := 0; j < int(C.BLOOM_FILTER_K); j++ {
            cResults1Indexable[j] = (*C.uint8_t)(C.CBytes(resps[2 * i].Results[j]))
            cResults2Indexable[j] = (*C.uint8_t)(C.CBytes(resps[2 * i + 1].Results[j]))
        }

        clientLock.Lock()
        start := time.Now()
        C.assembleQueryResponses_malicious((*C.client)(c),
                                       (**C.uint8_t)(cResults1),
                                       (**C.uint8_t)(cResults2),
                                       (*C.uint32_t)(cIndexes[i]),
                                       (*C.uint8_t)(cDocsPresent))
        elapsed := time.Since(start)
        clientLock.Unlock()
        logLatency(elapsed, "reconstruct_malicious")


        docsPresent = append(docsPresent, C.GoBytes(cDocsPresent, C.int(C.NUM_DOCS_BYTES))...)
 
        C.free(cResults1)
        C.free(cResults2)
        C.free(cDocsPresent)
    }

    return docsPresent, nil
}

/* Run search (semihonest adversaries). */
func SearchKeyword_semihonest(conn *common.Conn, keyword string, useMaster bool) ([]byte, error) {
    if (useMaster) {
        GetState(conn);
    }
    
    // In progress
    cKeys1 := (**C.uchar)(C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&keyword))))
    cKeys2 := (**C.uchar)(C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&keyword))))
    cResults1 := C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&keyword)))
    cResults2 := C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof(&keyword)))
    cIndexes := C.malloc(C.size_t(C.BLOOM_FILTER_K) * C.size_t(unsafe.Sizeof((C.uint32_t)(1))))
    cDocsPresent := C.malloc(C.size_t(C.NUM_DOCS_BYTES))
    cKeys1Indexable := (*[1<<30 - 1]*C.uchar)(unsafe.Pointer(cKeys1))
    cKeys2Indexable := (*[1<<30 - 1]*C.uchar)(unsafe.Pointer(cKeys2))
    cResults1Indexable := (*[1<<30 - 1]*C.uint8_t)(unsafe.Pointer(cResults1))
    cResults2Indexable := (*[1<<30 - 1]*C.uint8_t)(unsafe.Pointer(cResults2))
   
    clientLock.Lock()
    C.generateKeywordQuery((*C.client)(c),
                            C.CString(keyword),
                            (**C.uchar)(cKeys1),
                            (**C.uchar)(cKeys2),
                            (*C.uint32_t)(cIndexes))
    clientLock.Unlock()

    keys1 := make([][]byte, int(C.BLOOM_FILTER_K))
    keys2 := make([][]byte, int(C.BLOOM_FILTER_K))

    for i := 0; i < int(C.BLOOM_FILTER_K); i++ {
        keys1[i] = C.GoBytes(unsafe.Pointer(cKeys1Indexable[i]), C.int(C.getDPFKeyLen()))
        keys2[i] = C.GoBytes(unsafe.Pointer(cKeys2Indexable[i]), C.int(C.getDPFKeyLen()))
    }
    log.Println("version: ", versionNum)
    req1 := &common.SearchRequest_semihonest{
        Keys: keys1,
        Version: versionNum,
    }
    req2 := &common.SearchRequest_semihonest{
        Keys: keys2,
        Version: versionNum,
    }

    var wg sync.WaitGroup
    wg.Add(2)

    resp1 := &common.SearchResponse_semihonest{}
    var respError1 error
    go func() {
        defer wg.Done()
        common.SendMessage(
            config.Addr[0] + config.Port[0],
            common.SEARCH_REQUEST_SEMIHONEST,
            req1,
            resp1,
            &respError1,
        )
    }()

    resp2 := &common.SearchResponse_semihonest{}
    var respError2 error
    go func() {
        defer wg.Done()
        common.SendMessage(
            config.Addr[1] + config.Port[1],
            common.SEARCH_REQUEST_SEMIHONEST,
            req2,
            resp2,
            &respError2,
        )
    }()
   
    wg.Wait()

    for i := 0; i < int(C.BLOOM_FILTER_K); i++ {
       cResults1Indexable[i] = (*C.uint8_t)(C.CBytes(resp1.Results[i]))
       cResults2Indexable[i] = (*C.uint8_t)(C.CBytes(resp2.Results[i]))
    }
    
    clientLock.Lock()
    C.assembleQueryResponses((*C.client)(c),
                            (**C.uint8_t)(cResults1),
                            (**C.uint8_t)(cResults2),
                            (*C.uint32_t)(cIndexes),
                            (*C.uint8_t)(cDocsPresent))
    clientLock.Unlock()


    docsPresent := C.GoBytes(cDocsPresent, C.int(C.NUM_DOCS_BYTES))

    return docsPresent, nil
}

func RunFastSetup(benchmarkDir string, useMaster bool) error {
    req1 := &common.SetupRequest{
        BenchmarkDir: benchmarkDir,
    }
    req2 := &common.SetupRequest{
        BenchmarkDir: benchmarkDir,
    }

    var wg sync.WaitGroup
    wg.Add(2)

    resp1 := &common.SetupResponse{}
    var respError1 error
    go func() {
        defer wg.Done()
        common.SendMessage(
            config.Addr[0] + config.Port[0],
            common.SETUP_REQUEST,
            req1,
            resp1,
            &respError1,
        )
    }()

    resp2 := &common.SetupResponse{}
    var respError2 error
    go func() {
        defer wg.Done()
        common.SendMessage(
            config.Addr[1] + config.Port[1],
            common.SETUP_REQUEST,
            req2,
            resp2,
            &respError2,
        )
    }()

    wg.Wait()

    if (useMaster) {
        req3 := &common.MasterSetupRequest{
            NumDocs:    resp1.NumDocs,
            Versions:   resp1.Versions,
        }
        resp3 := &common.MasterSetupResponse{}
        var respError3 error
        common.SendMessage(
            config.MasterAddr + config.MasterPort,
            common.SETUP_REQUEST,
            req3,
            resp3,
            &respError3,
        )
    }

    cVersions := C.malloc(C.size_t(C.MAX_DOCS) * 4)
    cVersionsIndexable := (*[1<<30 - 1]C.uint32_t)(unsafe.Pointer(cVersions))

    for i := 0; i < int(C.MAX_DOCS); i++ {
        cVersionsIndexable[i] = (C.uint32_t)(resp1.Versions[i])
    }

    clientLock.Lock()
    C.updateClientState((*C.client)(c),
                        (C.int)(resp1.NumDocs),
                        (*C.uint32_t)(cVersions))
    clientLock.Unlock()

    return nil
}

/* Run fast setup, possibly with multiple clusters (only for benchmarking/testing). */
func RunFastSetup_parallel(benchmarkDir string, useMaster bool, numClusters int) error {
    req := &common.SetupRequest{
        BenchmarkDir: benchmarkDir,
    }

    var wg sync.WaitGroup
    wg.Add(2 * numClusters)

    resp := make([]*common.SetupResponse, numClusters * 2)
    for i := 0; i < numClusters * 2; i++ {
        log.Println(i)
        resp[i] = &common.SetupResponse{}
        go func(i int) {
            defer wg.Done()
            var respError error
            common.SendMessage(
                config.Addr[i] + config.Port[i],
                common.SETUP_REQUEST,
                req,
                resp[i],
                &respError,
            )
        }(i)
    }

    wg.Wait()

    if (useMaster) {
        req3 := &common.MasterSetupRequest{
            NumDocs:    resp[0].NumDocs,
            Versions:   resp[0].Versions,
        }
        resp3 := &common.MasterSetupResponse{}
        var respError3 error
        common.SendMessage(
            config.MasterAddr + config.MasterPort,
            common.SETUP_REQUEST,
            req3,
            resp3,
            &respError3,
        )
    }

    cVersions := C.malloc(C.size_t(C.MAX_DOCS) * 4)
    cVersionsIndexable := (*[1<<30 - 1]C.uint32_t)(unsafe.Pointer(cVersions))

    for i := 0; i < int(C.MAX_DOCS); i++ {
        cVersionsIndexable[i] = (C.uint32_t)(resp[0].Versions[i])
    }

    clientLock.Lock()
    C.updateClientState((*C.client)(c),
                        (C.int)(resp[0].NumDocs),
                        (*C.uint32_t)(cVersions))
    clientLock.Unlock()

    return nil
}

/* Get state from master. */
func GetState(conn *common.Conn) error {
    req := &common.GetStateRequest{}
    resp := &common.GetStateResponse{}
    var respError error

    common.SendMessageWithConnection(
        conn,
        common.GET_STATE_REQUEST,
        req,
        resp,
        &respError,
    )

    /*cVersions := C.malloc(C.size_t(C.MAX_DOCS) * 4)
    cVersionsIndexable := (*[1<<30 - 1]C.uint32_t)(unsafe.Pointer(cVersions))

    for i := 0; i < int(C.MAX_DOCS); i++ {
        cVersionsIndexable[i] = (C.uint32_t)(resp.Versions[i])
    }*/

    clientLock.Lock()
    C.updateClientState((*C.client)(c),
                        (C.int)(resp.NumDocs),
                        nil)
    clientLock.Unlock()


    versionNum = resp.SysVersion
    return nil
}

func OpenConnection() *common.Conn {
    conn, err := common.OpenConnection(config.MasterAddr +  config.MasterPort)
    if err != nil {
        log.Fatalln("Error opening connection to master: ", err)
        return nil
    }
    return conn
}

func CloseConnection(conn *common.Conn) {
    common.CloseConnection(conn)
}

func Setup(configFile string, bloomFilterSz int, numFiles int) string {
    var err error
    config, err = setupConfig(configFile)
    if err != nil {
        log.Fatalln("Error retrieving config file: ", err)
        return ""
    }
    c = &C.client{}
    C.setSystemParams(C.int(bloomFilterSz), C.int(numFiles));
    maskKey := make([]byte, hex.DecodedLen(len(config.MaskKey)))
    macKey := make([]byte, hex.DecodedLen(len(config.MacKey)))
    hex.Decode(maskKey, []byte(config.MaskKey))
    hex.Decode(macKey, []byte(config.MacKey))
    C.initializeClient((*C.client)(c), C.int(kNumThreads), (*C.uint8_t)(C.CBytes(maskKey)), (*C.uint8_t)(C.CBytes(macKey)))
    C.initStemmer()
    return config.OutDir + strconv.Itoa(numFiles) + "_docs_" + strconv.Itoa(int(C.BLOOM_FILTER_SZ)) + "_latency"
}

func Cleanup() {
    C.freeClient((*C.client)(c))
}
