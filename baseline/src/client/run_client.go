package main

import(
    "flag"
    "oram"
    "fmt"
    "math/rand"
    "time"
)

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

var numAccessesMap = map[int]int {
    1024: 11,
    2048: 21,
    4096: 41,
    8192: 77,
    16384: 116,
    32768: 208,
    65536: 489,
    131072: 931,
    262144: 1836,
}

const avgNumKeywordsPerDoc = 73
const numTrials = 100

func runSearch(c *oram.Client, numAccesses int, numBlocks int) {
    block := make([]byte, blockSize)
    cm := make([]byte, 4 * numAccesses)
    c.LoadState()
    c.CommitToReq(cm)
    for j := 0; j < numAccesses; j++ {
        c.Access(false, rand.Intn(numBlocks), block)
    }
    c.SaveState()
}

func runUpdate(c *oram.Client, numAccesses int, numBlocks int) {
    block := make([]byte, blockSize)
    cm := make([]byte, 4 * numAccesses * avgNumKeywordsPerDoc)
    c.LoadState()
    c.CommitToReq(cm)
    for i := 0; i < avgNumKeywordsPerDoc; i++ {
        for j := 0; j < numAccesses; j++ {
            c.Access(false, rand.Intn(numBlocks), block)
        }
    }
    c.SaveState()
}

func main() {
    n := flag.Int("n", 1024, "number of docs")
    addr := flag.String("addr", "127.0.0.1:4441", "server IP addr and port")
    flag.Parse()
    numBlocks := numBlocksMap[*n]
    numAccesses := numAccessesMap[*n]

    c := oram.InitClient(numBlocks, z, blockSize)
    c.AddServer(*addr, numBlocks, z)
    fmt.Println("finished adding server")
    c.SaveState()

    numTrials := 10
/*
    oram.StartCtr()
    for i := uint64(0); i < numTrials; i++ {
        runSearch(c, numAccesses, numBlocks)
    }
    searchBW := oram.GetCtr()

    oram.StartCtr()
    for i := uint64(0); i < numTrials; i++ {
        runUpdate(c, numAccesses, numBlocks)
    }
    updateBW := oram.GetCtr()

    fmt.Printf("Search bw: %d KB\n", searchBW / 1024 / numTrials)
    fmt.Printf("Update bw: %d KB\n", updateBW / 1024 / numTrials)
*/
    startSearch := time.Now()
    for i := 0; i < numTrials; i++ {
        runSearch(c, numAccesses, numBlocks)
    }
    searchTime := time.Since(startSearch)

    fmt.Printf("Total time to search: %s\n", searchTime)
    fmt.Printf("Time to search: %s\n", searchTime / time.Duration(numTrials))

    startThroughput_1_9 := time.Now()
    for i := 0; i < numTrials / 10; i++ {
        runUpdate(c, numAccesses, numBlocks)
        for j := 0; j < 9; j++ {
            runSearch(c, numAccesses, numBlocks)
        }
    }
    throughputTime_1_9 := time.Since(startThroughput_1_9)

    fmt.Printf("Total time for 10/90 workload: %s\n", throughputTime_1_9)
    fmt.Printf("Throughput for 10/90 workload: %f ops/sec\n", float64(numTrials) / throughputTime_1_9.Seconds())

    startThroughput_5_5 := time.Now()
    for i := 0; i < numTrials / 10; i++ {
        for j := 0; j < 5; j++ {
            runUpdate(c, numAccesses, numBlocks)
        }
        for j := 0; j < 5; j++ {
            runSearch(c, numAccesses, numBlocks)
        }
    }
    throughputTime_5_5 := time.Since(startThroughput_5_5)

    fmt.Printf("Total time for 50/50 workload: %s\n", throughputTime_5_5)
    fmt.Printf("Throughput for 50/50 workload: %f ops/sec\n", float64(numTrials) / throughputTime_5_5.Seconds())

    startThroughput_9_1 := time.Now()
    for i := 0; i < numTrials / 10; i++ {
        for j := 0; j < 9; j++ {
            runUpdate(c, numAccesses, numBlocks)
        }
        runSearch(c, numAccesses, numBlocks)
    }
    throughputTime_9_1 := time.Since(startThroughput_9_1)

    fmt.Printf("Total time for 90/10 workload: %s\n", throughputTime_9_1)
    fmt.Printf("Throughput for 90/10 workload: %f ops/sec\n", float64(numTrials) / throughputTime_9_1.Seconds())

}
