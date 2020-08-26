package oram 

import(
    "net"
    "crypto/tls"
    "fmt"
    "bufio"

    "github.com/hashicorp/go-msgpack/codec"
)

type ConnectionHandler func(net.Conn)

const kThreads = 50 

type Conn struct {
    tlsConn     *tls.Conn
    r           *bufio.Reader
    w           *bufio.Writer
    dec         *codec.Decoder
    enc         *codec.Encoder
}

type ReadRequest struct {
    L   int
    N   int
}

type ReadResponse struct {
    BucketResp  Bucket
}

type WriteRequest struct {
    L           int
    N           int
    BucketReq   Bucket
}

type SaveRequest struct {
    StashCts    [][]byte
    PosMapCt    []byte
}

type LoadRequest struct {}

type LoadResponse struct {
    StashCts    [][]byte
    PosMapCt    []byte
}

type CommitRequest struct {
    Cm      []byte
}

const (
    READ uint8 = iota
    WRITE
    SAVE
    LOAD
    COMMIT
)

var byteCounter uint64 

func StartCtr() {
    byteCounter = 0
}

func GetCtr() uint64 {
    return byteCounter
}

func OpenConnection(dstAddr string) (*Conn, error) {
     conf := &tls.Config{
        InsecureSkipVerify: true,
    }

    conn := &Conn{}

    var err error
    conn.tlsConn, err = tls.Dial("tcp", dstAddr, conf)
    if err != nil {
        return nil, err
    }

    conn.r = bufio.NewReader(conn.tlsConn)
    conn.w = bufio.NewWriter(conn.tlsConn)
    conn.dec = codec.NewDecoder(conn.r, &codec.MsgpackHandle{})
    conn.enc = codec.NewEncoder(conn.w, &codec.MsgpackHandle{})

    return conn, err
}

func CloseConnection(conn *Conn) {
    conn.tlsConn.Close()
}

func SendMessageWithConnectionNoResp(conn *Conn, reqType uint8, req interface{}) error {
    if err := conn.w.WriteByte(reqType); err != nil {
        return err
    }

    if err := conn.enc.Encode(req); err != nil {
        return err
    }

    byteCounter += uint64(conn.w.Buffered())

    if err := conn.w.Flush(); err != nil {
        return err
    }

    return nil
}

func SendMessageWithConnection(conn *Conn, reqType uint8, req interface{}, resp interface{}, respErr *error) error {
    if err := conn.w.WriteByte(reqType); err != nil {
        return err
    }

    if err := conn.enc.Encode(req); err != nil {
        return err
    }

    byteCounter += uint64(conn.w.Buffered())

    if err := conn.w.Flush(); err != nil {
        return err
    }

    var errReq error
    if err := conn.dec.Decode(&errReq); err != nil {
        return err
    }

    byteCounter += uint64(conn.r.Buffered())

    if err := conn.dec.Decode(resp); err != nil {
        return err
    }

    return nil
}

func SendMessage(dstAddr string, reqType uint8, req interface{}, resp interface{}, respErr *error) error {
    conn, err := OpenConnection(dstAddr)
    if err != nil {
        return err
    }
    defer CloseConnection(conn)
    return SendMessageWithConnection(conn, reqType, req, resp, respErr)
}

func SendMessageNoResp(dstAddr string, reqType uint8, req interface{}) error {
    conn, err := OpenConnection(dstAddr)
    if err != nil {
        return err
    }
    defer CloseConnection(conn)
    return SendMessageWithConnectionNoResp(conn, reqType, req)
}

func ListenForSingleMessage(dstPort string, certFile string, keyFile string, handler ConnectionHandler) error {
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return err
    }

    conf := &tls.Config{Certificates: []tls.Certificate{cert}}
    ln, err := tls.Listen("tcp", dstPort, conf)
    if err != nil {
        return err
    }

    conn, err := ln.Accept()
    if err != nil {
        return err
    }
    handler(conn)
    return nil
}

func ListenLoop(dstPort string, certFile string, keyFile string, handler ConnectionHandler) error {
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return err
    }

    conf := &tls.Config{Certificates: []tls.Certificate{cert}}
    ln, err := tls.Listen("tcp", dstPort, conf)
    if err != nil {
        return err
    }

    /*jobsChan := make(chan func(), kThreads)
    for i := 0; i < kThreads; i++ {
        go func() {
            for {
                job := <-jobsChan
                job()
            }
        }()
    }*/

    for {
        conn, err := ln.Accept()
        if err != nil {
            fmt.Println(err)
            continue
        }
        // TODO: need real synchronization in index!
//        go handler(conn)
        handler(conn)
        //jobsChan <- func() {handler(conn)}
    }
    return nil
}
