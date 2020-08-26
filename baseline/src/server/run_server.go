package main

import(
    "flag"
    "log"
    "bufio"
    "oram"
    "github.com/hashicorp/go-msgpack/codec"
    "net"
    "fmt"
)

const certFile = "../src/config/server1.crt"
const keyFile = "../src/config/server1.key"

var s *oram.Server

var stashCts [][]byte
var posMapCt []byte
var cm []byte

const blockSize = 100 * 4
const z = 4

var numBlocksMap = map[int]int{
    1024: 16479,
    2048: 18187,
    4096: 20220,
    8192: 25782,
    16384: 37739,
    32768: 69029,
    65536: 104213,
    131072: 179095,
    262144: 344721,
}

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
                //log.Println("Connection with", conn.RemoteAddr(), "closed")
                break
            } else {
                log.Fatalln(err)
            }   
        }   

        switch rpcType {

        case oram.READ:
            var req oram.ReadRequest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            bucket, respErr := s.ReadNode(req.L, req.N)
            resp := oram.ReadResponse{
                BucketResp: bucket,
            }
            if err := enc.Encode(&respErr); err != nil {
                log.Fatalln(err)
            }
            if err := enc.Encode(&resp); err != nil {
                log.Fatalln(err)
            }
            if err := w.Flush(); err != nil {
                log.Fatalln(err)
            }

        case oram.WRITE:
            var req oram.WriteRequest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            s.WriteNode(req.BucketReq, req.L, req.N)

        case oram.SAVE:
            var req oram.SaveRequest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            stashCts = make([][]byte, 0)
            for i := 0; i < len(req.StashCts); i++ {
                stashCts = append(stashCts, req.StashCts[i])
            }
            posMapCt = make([]byte, len(req.PosMapCt))
            copy(posMapCt, req.PosMapCt)

        case oram.LOAD:
            var req oram.LoadRequest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            var respErr error
            resp := oram.LoadResponse{
                StashCts: stashCts,
                PosMapCt: posMapCt,
            }
            if err := enc.Encode(&respErr); err != nil {
                log.Fatalln(err)
            }
            if err := enc.Encode(&resp); err != nil {
                log.Fatalln(err)
            }
            if err := w.Flush(); err != nil {
                log.Fatalln(err)
            }

        case oram.COMMIT:
            var req oram.CommitRequest
            if err := dec.Decode(&req); err != nil {
                log.Fatalln(err)
            }
            cm = make([]byte, len(req.Cm))
            copy(cm, req.Cm)

        default:
            log.Fatalln(fmt.Errorf("Unknown request type %d", rpcType))
        }
    }
}

func main() {
    n := flag.Int("n", 1024, "number of documents")
    port := flag.String("port", ":4441", "number of blocks in each bucket")
    flag.Parse()
    s = oram.InitServer(numBlocksMap[*n], z, blockSize)
    err := oram.ListenLoop(*port, certFile, keyFile, handleConnection)
    if err != nil {
        log.Fatalln(err)
    }
}
