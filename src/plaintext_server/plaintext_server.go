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
)

var config common.ServerConfig

var index map[string][]int

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
func searchKeyword_plaintext(req common.SearchRequest_plaintext) (common.SearchResponse_plaintext, error) {

    resp := common.SearchResponse_plaintext{
        Results: index[req.Keyword],
    }

    return resp, nil
}

/* Only called directly if not using the master; for updating documents (malicious adversaries). */
func updateDoc_plaintext(req common.UpdateRequest_plaintext) (common.UpdateResponse_plaintext, error) {


    for _, keyword := range req.Keywords {
        if _, ok := index[keyword]; ok {
            for _, docID := range index[keyword] {
                if (docID == req.DocID) {
                    continue
                }
            }
            index[keyword] = append(index[keyword], req.DocID)
        } else {
            index[keyword] = []int{req.DocID}
        }
    }

    return common.UpdateResponse_plaintext{Test: "success"}, nil

}


func main() {
    /* Set up config */
    filename := flag.String("config", "src/config/server1.config", "server config file")
    flag.Parse()

    index = make(map[string][]int)

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
