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
    "bufio"
    "github.com/hashicorp/go-msgpack/codec"
    "flag"
    "time"
    "strconv"
    "io"
    "unsafe"
)

var config common.ServerConfig

var newIndex map[string][]int
var oldIndex map[string][]int
var incomingIndex map[string][]int

var oldVersionNum int
var newVersionNum int
var incomingVersionNum int

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

        case common.SEARCH_REQUEST_PLAINTEXT:
            var req common.SearchRequest_plaintext
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            start := time.Now()
            resp, respErr := searchKeyword_plaintext(req)
            elapsed := time.Since(start)
            resp.ServerLatency = elapsed
            logLatency(elapsed, "plaintext")
            if err := enc.Encode(&respErr); err != nil {
                log.Fatalln(err)
            }
            if err := enc.Encode(&resp); err != nil {
                log.Fatalln(err)
            }
            if err := w.Flush(); err != nil {
                log.Fatalln(err)
            }


        case common.UPDATE_REQUEST_PLAINTEXT:
            var req common.UpdateRequest_plaintext
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            updateDoc_plaintext(req)


        case common.INDEX_SZ_REQUEST:
            var req common.IndexSzRequest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            resp, respErr := getIndexSize()
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

func logLatency (latency time.Duration, tag string) {
    go func(latency time.Duration, tag string) {
        file, _ := os.Create(config.OutDir + "/" + strconv.Itoa(int(C.NUM_DOCS)) + "_docs_" + strconv.Itoa(int(C.BLOOM_FILTER_SZ)) + "bf_latency_" + tag)
        defer file.Close()
        io.WriteString(file, latency.String())
    }(latency, tag)
}

func getIndexSize() (common.IndexSzResponse, error) {
    sz := 0
    for keyword := range(newIndex) {
        sz += len(keyword)
        sz += len(newIndex[keyword]) * int(unsafe.Sizeof(newIndex[keyword][0]))
    }
    resp := common.IndexSzResponse{
        Size: sz,
    }
    return resp, nil
}

/* Process search request (malicious adversaries). */
func searchKeyword_plaintext(req common.SearchRequest_plaintext) (common.SearchResponse_plaintext, error) {

    var index map[string][]int
    if (req.Version == newVersionNum) {
        index = newIndex
    } else {
        index = oldIndex
    }
    resp := common.SearchResponse_plaintext{
        Results: index[req.Keyword],
    }

    return resp, nil
}

/* Only called directly if not using the master; for updating documents (malicious adversaries). */
func updateDoc_plaintext(req common.UpdateRequest_plaintext) (common.UpdateResponse_plaintext, error) {


    for _, keyword := range req.Keywords {
        if _, ok := newIndex[keyword]; ok {
            for _, docID := range newIndex[keyword] {
                if (docID == req.DocID) {
                    continue
                }
            }
            newIndex[keyword] = append(newIndex[keyword], req.DocID)
        } else {
            newIndex[keyword] = []int{req.DocID}
        }
    }

    return common.UpdateResponse_plaintext{Test: "success"}, nil

}

/* Start processing incoming batch of updates from master. */
func startUpdateBatch(req common.BatchStartRequest) (common.BatchStartResponse, error) {
    incomingIndex = req.PlaintextUpdates
    incomingVersionNum = req.VersionNum
    return common.BatchStartResponse{Commit: true}, nil
}

/* Finish processing incoming batch of updates from master. */
func finishUpdateBatch(req common.BatchFinishRequest) (common.BatchFinishResponse, error) {
    for k,v := range newIndex {
        newIndex[k] = v
    }
    oldVersionNum =  newVersionNum
    /* Process all incoming updates. */


    for keyword := range incomingIndex {
        if _, ok := newIndex[keyword]; ok {
            for _,doc := range(incomingIndex[keyword]) {
                present := false
                for _, currDoc := range(newIndex[keyword]) {
                    if doc == currDoc {
                        present = true
                    }
                }
                if !present {
                    newIndex[keyword] = append(newIndex[keyword], doc)
                }
            }
        } else {
            newIndex[keyword] = incomingIndex[keyword] 
        }
    }

    newVersionNum = incomingVersionNum
    return common.BatchFinishResponse{}, nil
}


func main() {
    /* Set up config */
    filename := flag.String("config", "src/config/server1.config", "server config file")
    flag.Parse()

    newIndex = make(map[string][]int)
    oldIndex = make(map[string][]int)
    oldVersionNum = 0
    newVersionNum = 0

    var err error
    config, err = setupConfig(*filename)
    if err != nil {
        log.Fatalln("Error retrieving config file: ", err)
    }

    /* Start listening */
    log.Println("Listening...")
    err = common.ListenLoop(config.Port, config.CertFile, config.KeyFile, handleConnection)
    if err != nil {
        log.Fatalln(err)
    }
}
