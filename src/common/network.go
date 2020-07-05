package common

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

/* Open a TLS connection to IP addr and port pair. */
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

/* Close a TLS connection. */
func CloseConnection(conn *Conn) {
    conn.tlsConn.Close()
}

/* Send a message over a connection (no response expected). */
func SendMessageWithConnectionNoResp(conn *Conn, reqType uint8, req interface{}) error {
    if err := conn.w.WriteByte(reqType); err != nil {
        return err
    }

    if err := conn.enc.Encode(req); err != nil {
        return err
    }

    if err := conn.w.Flush(); err != nil {
        return err
    }

    return nil
}

/* Send a message over a connection (wait for a response). */
func SendMessageWithConnection(conn *Conn, reqType uint8, req interface{}, resp interface{}, respErr *error) error {
    if err := conn.w.WriteByte(reqType); err != nil {
        return err
    }

    if err := conn.enc.Encode(req); err != nil {
        return err
    }

    if err := conn.w.Flush(); err != nil {
        return err
    }

    var errReq error
    if err := conn.dec.Decode(&errReq); err != nil {
        return err
    }
    if err := conn.dec.Decode(resp); err != nil {
        return err
    }

    return nil
}

/* Send a message, opening and closing a connection (wait for a response). */
func SendMessage(dstAddr string, reqType uint8, req interface{}, resp interface{}, respErr *error) error {
    conn, err := OpenConnection(dstAddr)
    if err != nil {
        return err
    }
    defer CloseConnection(conn)
    return SendMessageWithConnection(conn, reqType, req, resp, respErr)
}

/* Send a message, opening and closing a connection  (no response expected). */
func SendMessageNoResp(dstAddr string, reqType uint8, req interface{}) error {
    conn, err := OpenConnection(dstAddr)
    if err != nil {
        return err
    }
    defer CloseConnection(conn)
    return SendMessageWithConnectionNoResp(conn, reqType, req)
}

/* Listen for a single message and run a handler to respond. */
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

/* Loop listening for incoming connections. */
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

    for {
        conn, err := ln.Accept()
        if err != nil {
            fmt.Println(err)
            continue
        }
        go handler(conn)
    }
    return nil
}
