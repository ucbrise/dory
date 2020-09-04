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
import(
    "encoding/json"
    "sync"
    "time"
    "common"
    "os"
    "net"
    "bufio"
    "log"
    "github.com/hashicorp/go-msgpack/codec"
    "fmt"
    "strconv"
    "io"
    "flag"
)

var config common.MasterConfig

var updateList []common.Update
var updateLocks []sync.Mutex

var versionNum int
var versionNumLock sync.Mutex

var maxDocs int
var docVersions []uint32
var numDocs int
var isMalicious bool

var tickTime time.Duration

var numClusters int

/* Read in config file. */
func setupConfig(filename string) (common.MasterConfig, error) {
    config := common.MasterConfig{}
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

/* Start two phase commit with servers for cached updates. */
func start2PC(version int, updateMapCopy map[int]common.Update) {

    commitReq := &common.BatchStartRequest{
        VersionNum:     version,
	    Updates:	updateMapCopy,
        Malicious:      isMalicious, 
    }

    finalCommit := true
    var finalCommitLock sync.Mutex

    var wg1 sync.WaitGroup
    wg1.Add(2 * numClusters)

    for i := 0; i < 2 * numClusters; i+= 1 {
        go func(i int) {
            var commitRespError error
            commitResp := &common.BatchStartResponse{}
            defer wg1.Done()
            common.SendMessage(
                config.Addr[i] + config.Port[i],
                common.BATCH_START_REQUEST,
                commitReq,
                commitResp,
                &commitRespError,
            )
            finalCommitLock.Lock()
            finalCommit = commitResp.Commit
            finalCommitLock.Unlock()
        }(i)
    }

    wg1.Wait()

    finishReq := &common.BatchFinishRequest{
        Commit:     finalCommit,
    }

    var wg2 sync.WaitGroup
    wg2.Add(2 * numClusters)

    for i := 0; i < 2 * numClusters; i += 1 {
        go func(i int) {
            finishResp := &common.BatchFinishResponse{}
            var finishRespError error
            defer wg2.Done()
            common.SendMessage(
                config.Addr[i] + config.Port[i],
                common.BATCH_FINISH_REQUEST,
                finishReq,
                finishResp,
                &finishRespError,
            )
        }(i)
    }

    wg2.Wait()

}

/* Store and aggregate update (semihonest adversaries). */
func updateDoc_semihonest(req common.UpdateRequest_semihonest) (common.UpdateResponse_semihonest, error) {
    if req.DocID >= numDocs {
        numDocs = req.DocID + 1
    }
    updateLocks[req.DocID].Lock()
    updateList[req.DocID] = common.Update{
        BF: req.BF,
        MACs: nil,
    }
    docVersions[req.DocID] = req.Version
    updateLocks[req.DocID].Unlock()
    return common.UpdateResponse_semihonest{}, nil
}

/* Store and aggregate update (malicious adversaries). */
func updateDoc_malicious(req common.UpdateRequest_malicious) (common.UpdateResponse_malicious, error) {
    if req.DocID >= numDocs {
        numDocs = req.DocID + 1
    }

    updateLocks[req.DocID].Lock()
    updateList[req.DocID] = common.Update  {
        BF: req.BF,
        MACs: req.MACs,
    }
    docVersions[req.DocID] = req.Version
    updateLocks[req.DocID].Unlock()

    return common.UpdateResponse_malicious{}, nil
}

/* Process getState request by returning current version number. */
func getState (req common.GetStateRequest) (common.GetStateResponse, error) {
    docVersionsCopy := make([]uint32, len(docVersions))
    copy(docVersionsCopy, docVersions)
    versionNumLock.Lock()
    resp := common.GetStateResponse{
        NumDocs: numDocs,
        SysVersion: versionNum,
        //Versions: docVersionsCopy,
    }
    versionNumLock.Unlock()
    return resp, nil
}

func testSetup (req common.MasterSetupRequest) (common.MasterSetupResponse, error) {
    numDocs = req.NumDocs
    docVersions = req.Versions
    return common.MasterSetupResponse{Test: "success"}, nil
}

/* Dispatch handlers for different message types. */
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

        case common.UPDATE_REQUEST_SEMIHONEST:
            var req common.UpdateRequest_semihonest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            updateDoc_semihonest(req)

        case common.UPDATE_REQUEST_MALICIOUS:
            var req common.UpdateRequest_malicious
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            updateDoc_malicious(req)

        case common.SETUP_REQUEST:
            var req common.MasterSetupRequest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            resp, respErr := testSetup(req)
            if err := enc.Encode(&respErr); err != nil {
                log.Fatalln(err)
            }
            if err := enc.Encode(&resp); err != nil {
                log.Fatalln(err)
            }
            if err := w.Flush(); err != nil {
                log.Fatalln(err)
            }

        case common.GET_STATE_REQUEST:
            var req common.GetStateRequest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            resp, respErr := getState(req)
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

/* Periodically run 2PC with servers to send cached updates and increment system-wide
 * version number. */
func tickLoop() {
    tick := time.Tick(tickTime)
    for {
        <-tick
        if len(updateList) == 0 {
            continue
        }
	    updateMapCopy := make(map[int]common.Update)
        for docID := 0; docID < len(updateList); docID++ {
            updateLocks[docID].Lock()
            updateMapCopy[docID]  = updateList[docID]
        }
        listLen := len(updateList)
        updateList = make([]common.Update, listLen)
        for docID := 0; docID < len(updateList); docID++ {
            updateLocks[docID].Unlock()
        }
        start2PC(versionNum + 1, updateMapCopy)
        versionNumLock.Lock()
        versionNum += 1
        versionNumLock.Unlock()
    }
}


func main() {
/* Set up config */
    filename := flag.String("config", "src/config/server1.config", "server config file")
    maxDocsFlag := flag.Int("max_docs", 256, "max number of documents")
    bloomFilterSz := flag.Int("bf_sz", 128, "bloom filter size in bits")
    tickMs := flag.Int("tick_ms", 1000, "batch time in ms")
    isMaliciousFlag := flag.Bool("malicious", true, "run with malicious checks")
    numClustersFlag := flag.Int("num_clusters", 1, "number of clusters")
    flag.Parse()
    maxDocs = *maxDocsFlag
    isMalicious = *isMaliciousFlag
    numClusters = *numClustersFlag
    tickTime = time.Duration(*tickMs) * time.Millisecond
    log.Println("tick time: ", tickTime)

    var err error
    config, err = setupConfig(*filename)
    if err != nil {
        log.Fatalln("Error retrieving config file: ", err)
    }

    /* Initialize server */
    log.Println("Starting initialization...")
    C.setSystemParams(C.int(*bloomFilterSz), C.int(maxDocs));
    versionNum = 0
    docVersions =  make([]uint32, maxDocs)
    updateList = make([]common.Update, maxDocs)
    updateLocks = make([]sync.Mutex, maxDocs)

    go tickLoop()

    /* Start listening */
    log.Println("Listening...")
    err = common.ListenLoop(config.MasterPort, config.CertFile, config.KeyFile, handleConnection)
    if err != nil {
        log.Fatalln(err)
    }
}

