package common

import (
    "time"
)

const MAX_KEYWORD_SIZE = 32

type Server struct {
    Addr            string
    ID              string
    Port            string
    CertFile        string
    KeyFile         string
}

type SystemConfig struct {
    MasterAddr      string
    MasterID        string
    MasterPort      string
    MasterCertFile  string
    MasterKeyFile   string
    ClientAddrs     []string
    ClientIDs       []string
    Servers         []Server
    OutDir          string
    ClientMaskKey   string
    ClientMacKey    string
    SSHKeyPath      string
    BaselineServerAddr  string
    BaselineServerID    string
    BaselineClientAddr  string
    BaselineClientID    string
}

type MasterConfig struct {
    MasterAddr      string
    MasterPort      string
    Addr            []string
    Port            []string
    CertFile        string
    KeyFile         string
    OutDir          string
}

type ServerConfig struct {
    Addr            string
    Port            string
    CertFile        string
    KeyFile         string
    OutDir          string
    ClientMaskKey   string
    ClientMacKey    string
}

type ClientConfig struct {
    MasterAddr      string
    MasterPort      string
    Addr            []string
    Port            []string
    OutDir          string
    MaskKey         string
    MacKey          string
}

type Update struct {
    BF      []byte
    MACs    [][]byte
}

type SearchRequest_malicious struct {
    Keys        [][]byte
    Version     int
}

type SearchRequest_semihonest struct {
    Keys        [][]byte
    Version     int
}

type SearchResponse_malicious struct {
    Results         [][]byte
    ServerLatency   time.Duration
}

type SearchResponse_semihonest struct {
    Results         [][]byte
    ServerLatency   time.Duration
}

type UpdateRequest_malicious struct {
    DocID           int
    Version         uint32
    BF              []byte
    MACs            [][]byte
}

type UpdateRequest_semihonest struct {
    DocID           int
    Version         uint32
    BF              []byte
}

type UpdateResponse_malicious struct {
    Test            string
}

type UpdateResponse_semihonest struct {
    Test            string
}

type SetupRequest struct {
    BenchmarkDir    string

}

type SetupResponse struct {
    NumDocs         int
    Versions        []uint32
}

type MasterSetupRequest struct {
    NumDocs         int
    Versions        []uint32
}

type MasterSetupResponse struct {
    Test            string
}

type BatchStartRequest struct {
    VersionNum      int
    Updates         map[int]Update
    Malicious       bool
}

type BatchStartResponse struct {
    Commit          bool
}

type BatchFinishRequest struct {
    Commit          bool
}

type BatchFinishResponse struct {
}


type GetStateRequest struct {
    Test            string
}

type GetStateResponse struct {
    NumDocs         int
    SysVersion      int
//    Versions        []uint32
}

const (
    SEARCH_REQUEST_MALICIOUS uint8 = iota
    SEARCH_REQUEST_SEMIHONEST
    UPDATE_REQUEST_MALICIOUS
    UPDATE_REQUEST_SEMIHONEST
    SETUP_REQUEST
    GET_STATE_REQUEST
    BATCH_START_REQUEST
    BATCH_FINISH_REQUEST
)
