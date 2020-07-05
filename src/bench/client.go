package main

// TODO: figure out openmp and add back -fopenmp

import (
    "client"
    "log"
    "flag"
    "os"
    "io/ioutil"
    "io"
    "time"
    "sync"
    "strconv"
)

/* Run correctness tests. */
func correctnessTests(configFile string, bloomFilterSz int, numDocs int, isMalicious bool, useMaster bool) {
    /* Initialize client */
    log.Println("Initializing client...")
    client.Setup(configFile, bloomFilterSz, numDocs)

    log.Println("bloom filter sz: ", bloomFilterSz, ", num docs: ", numDocs)
    keywords := []string{"hello", "world"}
    conn := client.OpenConnection()
    for i := 0; i < numDocs; i++ {
        if (isMalicious) {
            client.UpdateDoc_malicious(conn, keywords, i, useMaster)
        } else {
            client.UpdateDoc_semihonest(conn, keywords, i, useMaster)
        }
    }
    time.Sleep(2000 * time.Millisecond)

    log.Println("Searching for 'hello'")
    var err error
    var results []byte
    if (isMalicious) {
        results, err, _, _, _, _ = client.SearchKeyword_malicious(conn, "hello", useMaster)
    } else {
        results, err = client.SearchKeyword_semihonest(conn, "hello", useMaster)
    }
    if err != nil {
        log.Fatalln(err)
    }
    passed := true
    for i := 0; i < len(results); i++ {
        if (results[i] != 0xff) {
            passed = false
        }
    }
    log.Println("results: ", results)
    if (passed) {
        log.Println("----- PASSED TEST -----")
    } else {
        log.Println("----- FAILED TEST -----")
    }


    /* Clean up */
    log.Println("Cleaning up...")
    client.CloseConnection(conn)
    client.Cleanup()
}

/* Measure search latency with parallelism. */
func runParallelBenchmark(configFile string, numDocs int, bloomFilterSz int, useMaster bool, numClusters int) {
    /* Initialize client */

    numTrials := 1

    log.Println("Initializing client...")
    outFile := client.Setup(configFile, bloomFilterSz, numDocs)
    conn := client.OpenConnection()

    client.RunFastSetup_parallel("", useMaster, numClusters);
    time.Sleep(2000 * time.Millisecond)

    log.Println("Finished updates")

    var err error
    start := time.Now()
    for i := 0; i < numTrials; i++ {
        client.SearchKeyword_malicious_parallel(conn, "hello", useMaster, 1)
    }

    elapsed := time.Since(start)
    if err != nil {
        log.Fatalln(err)
    }

    tag := "parallel" + string(numClusters)
    file, err := os.Create(outFile + "_" + tag)
    if err != nil {
        log.Println(err)
    }
    defer file.Close()
    timeMs := float64(elapsed.Nanoseconds())/float64(1e6)/float64(numTrials)
    _, err = io.WriteString(file, strconv.FormatFloat(timeMs, 'f', 3, 64))
    if err != nil {
        log.Println(err)
    }

    log.Printf("time to search: %f ms\n", timeMs);

    /* Clean up */
    log.Println("Cleaning up...")
    log.Printf("0 0 0 0 %f\n", timeMs)
    client.CloseConnection(conn)
    client.Cleanup()
}

/* Run search latency benchmark without parallelism, including time breakdown. */
func runArtificialBenchmark(configFile string, numDocs int, bloomFilterSz int, isMalicious bool, fastSetup bool, useMaster bool) {
    /* Initialize client */

    numTrials := 1

    log.Println("Initializing client...")
    outFile := client.Setup(configFile, bloomFilterSz, numDocs)
    conn := client.OpenConnection()

    if (fastSetup) {
        client.RunFastSetup("", useMaster);
    } else {
        for i := 0; i < numDocs; i++ {
            if (isMalicious) {
                client.UpdateDoc_malicious(conn, []string{"hello", "world"}, i, useMaster)
            } else {
                client.UpdateDoc_semihonest(conn, []string{"hello", "world"}, i, useMaster)
            }
        }
    }
    time.Sleep(2000 * time.Millisecond)

    log.Println("Finished updates")

    var err error
    start := time.Now()
    getStateMs := 0.0
    client1Ms := 0.0
    networkAndServerMs := 0.0
    client2Ms := 0.0
    for i := 0; i < numTrials; i++ {
        if (isMalicious) {
            _, _, t1, t2, t3, t4  := client.SearchKeyword_malicious(conn, "hello", useMaster)
            getStateMs += float64(t1.Nanoseconds())/float64(1e6)
            client1Ms += float64(t2.Nanoseconds())/float64(1e6)
            networkAndServerMs += float64(t3.Nanoseconds())/float64(1e6)
            client2Ms += float64(t4.Nanoseconds())/float64(1e6)
        } else {
            _, err = client.SearchKeyword_semihonest(conn, "hello", useMaster)
        }
    }
    getStateMs = getStateMs/float64(numTrials)
    client1Ms = client1Ms/float64(numTrials)
    networkAndServerMs = networkAndServerMs/float64(numTrials)
    client2Ms = client2Ms/float64(numTrials)

    elapsed := time.Since(start)
    if err != nil {
        log.Fatalln(err)
    }

    tag := ""
    if (isMalicious)  {
        tag = "malicious"
    } else {
        tag = "semihonest"
    }
    file, err := os.Create(outFile + "_" + tag)
    if err != nil {
        log.Println(err)
    }
    defer file.Close()
    timeMs := float64(elapsed.Nanoseconds())/float64(1e6)/float64(numTrials)
    _, err = io.WriteString(file, strconv.FormatFloat(timeMs, 'f', 3, 64))
    _, err = io.WriteString(file, strconv.FormatFloat(getStateMs, 'f', 3, 64))
    _, err = io.WriteString(file, strconv.FormatFloat(client1Ms, 'f', 3, 64))
    _, err = io.WriteString(file, strconv.FormatFloat(networkAndServerMs, 'f', 3, 64))
    _, err = io.WriteString(file, strconv.FormatFloat(client2Ms, 'f', 3, 64))
    if err != nil {
        log.Println(err)
    }

    log.Printf("time to search: %f ms\n", timeMs);
    log.Printf("time to get state: %f ms\n", getStateMs);
    log.Printf("time for first client ops: %f ms\n", client1Ms);
    log.Printf("time for network/server ops: %f ms\n", networkAndServerMs);
    log.Printf("time for seconds client ops: %f ms\n", client2Ms);

    /* Clean up */
    log.Println("Cleaning up...")
    log.Printf("%f %f %f %f %f\n", getStateMs, client1Ms, networkAndServerMs, client2Ms, timeMs)
    client.CloseConnection(conn)
    client.Cleanup()
}

/* Measure update latency using documents from directory. */
func runDirBenchmark(configFile string, benchmarkDir string, bloomFilterSz int, numDocs int, isMalicious bool, useMaster bool) {
    topDirs, err := ioutil.ReadDir(benchmarkDir)
    if err != nil {
        log.Fatal(err)
    }

    /* Initialize client */
    log.Println("Initializing client...")
    outFile := client.Setup(configFile, bloomFilterSz, numDocs)
    conn := client.OpenConnection()

    log.Println("did setup")

    ctr := 0
    start := time.Now()
    for _,topDir := range topDirs {
        midDirs, _ := ioutil.ReadDir(benchmarkDir + "/" + topDir.Name())
        for _,midDir := range midDirs {
            files,_ := ioutil.ReadDir(benchmarkDir + "/" + topDir.Name() + "/" + midDir.Name())
            for docID, file := range files {
                var err error
                docID = docID % numDocs
                filename := benchmarkDir + "/" + topDir.Name() + "/" + midDir.Name() + "/" + file.Name()
                if (isMalicious) {
                    err = client.UpdateDocFile_malicious(conn, filename, docID, useMaster)
                } else {
                    err = client.UpdateDocFile_semihonest(conn, filename, docID, useMaster)
                }
                ctr += 1
                if err != nil {
                    log.Fatal(err)
                }
            }
        }
    }
    elapsed := float64(time.Since(start).Nanoseconds())/float64(1e6)/float64(ctr)

    log.Println("Finished updates")

    file, err := os.Create(outFile + "_UPDATE")
    if err != nil {
        log.Println(err)
    }
    defer file.Close()
    _, err = io.WriteString(file, strconv.FormatFloat(elapsed, 'f', 3, 64))
    if err != nil {
        log.Println(err)
    }

    log.Printf("avg update time: %s ms\n", strconv.FormatFloat(elapsed, 'f', 3, 64))
    log.Printf("%s\n", strconv.FormatFloat(elapsed, 'f', 3, 64))

    /* Clean up */
    client.CloseConnection(conn)
    client.Cleanup()
}

/* Start fast setup at servers for benchmarking/testing. */
func runFastSetup(configFile string, numDocs int, bloomFilterSz int, useMaster bool, numClusters int) {
    log.Println("Initializing client...")
    client.Setup(configFile, bloomFilterSz, numDocs)

    client.RunFastSetup_parallel("", useMaster, numClusters);
}

/* Run throughput benchmarks with  mix of updates and searches with multiple clusters. */
func runThroughputClustersBenchmark(configFile string, numDocs int, bloomFilterSz int, useMaster bool, seconds int, threads int, numUpdates int, numSearches int, numClusters int) {
    /* Initialize client */
    log.Println("Initializing client...")
    outFile := client.Setup(configFile, bloomFilterSz, numDocs)

    duration := time.Duration(seconds) * time.Second
    var wg sync.WaitGroup
    wg.Add(threads)
    totals := make([]int, threads)
    slice := numDocs / threads

    for i := 0; i < threads; i++ {
        go func(index int) {
            defer wg.Done()
            tick := time.Tick(duration)
            conn := client.OpenConnection()
            defer client.CloseConnection(conn)
            j := index * slice
            updateCtr := 0
            searchCtr := numSearches    // start with updates
            for {
                select {
                case _, ok := <-tick:
                    if (ok) {
                        totals[index] = j - (index * slice)
                        return
                    }
                    log.Println("channel but not ok???")
                default:
                    if (updateCtr < numUpdates) {
                        client.DummyUpdateDoc_malicious(conn, []string{"hello", "world"}, (j % numDocs), useMaster)
                        updateCtr += 1
                        if (updateCtr == numUpdates) {
                            searchCtr = 0
                        }
                    } else if (searchCtr < numSearches) {
                        client.SearchKeyword_malicious_parallel(conn, "hello", useMaster, 1)
                        searchCtr += 1
                        if (searchCtr == numSearches) {
                            updateCtr = 0
                        }
                    }
                    j += 1
                }
            }
        }(i)
    }
    wg.Wait()
    totalUpdates := 0
    for _, total := range(totals) {
        totalUpdates += total
    }

    file, err := os.Create(outFile + "_mixed_throughput")
    if err != nil {
        log.Println(err)
    }
    defer file.Close()
    _, err = io.WriteString(file, string(totalUpdates))
    if err != nil {
        log.Println(err)
    }

    log.Printf("With %d threads running for %s, ran %d updates\n", threads, duration, totalUpdates)
    log.Printf("Updates/sec: %f\n", float64(totalUpdates) / duration.Seconds())
    log.Printf("%d\n", totalUpdates)

    /* Clean up */
    client.Cleanup()
}

/* Run throughput benchmarks with single cluster. */
func runThroughputBenchmark(configFile string, numDocs int, bloomFilterSz int, isMalicious bool, useMaster bool, seconds int, threads int, numUpdates int, numSearches int) {
    /* Initialize client */
    log.Println("Initializing client...")
    outFile := client.Setup(configFile, bloomFilterSz, numDocs)

    duration := time.Duration(seconds) * time.Second
    var wg sync.WaitGroup
    wg.Add(threads)
    totals := make([]int, threads)
    slice := numDocs / threads

    for i := 0; i < threads; i++ {
        go func(index int) {
            defer wg.Done()
            tick := time.Tick(duration)
            conn := client.OpenConnection()
            defer client.CloseConnection(conn)
            j := index * slice
            updateCtr := 0
            searchCtr := numSearches    // start with updates
            for {
                select {
                case _, ok := <-tick:
                    if (ok) {
                        totals[index] = j - (index * slice)
                        return
                    }
                    log.Println("channel but not ok???")
                default:
                    if (updateCtr < numUpdates) {
                        if (isMalicious) {
                            client.DummyUpdateDoc_malicious(conn, []string{"hello", "world"}, (j % numDocs), useMaster)
                        } else {
                            client.DummyUpdateDoc_semihonest(conn, []string{"hello", "world"}, (j % numDocs), useMaster)
                        }
                        updateCtr += 1
                        if (updateCtr == numUpdates) {
                            searchCtr = 0
                        }
                    } else if (searchCtr < numSearches) {
                        client.GetState(conn)
                        searchCtr += 1
                        if (searchCtr == numSearches) {
                            updateCtr = 0
                        }
                    }
                    j += 1
                }
            }
        }(i)
    }
    wg.Wait()
    totalUpdates := 0
    for _, total := range(totals) {
        totalUpdates += total
    }

    tag := ""
    if (isMalicious)  {
        tag = "malicious_numUupdates"
    } else {
        tag = "semihonest_numUpdates"
    }
    file, err := os.Create(outFile + "_" + tag)
    if err != nil {
        log.Println(err)
    }
    defer file.Close()
    _, err = io.WriteString(file, string(totalUpdates))
    if err != nil {
        log.Println(err)
    }

    log.Printf("With %d threads running for %s, ran %d updates\n", threads, duration, totalUpdates)
    log.Printf("Updates/sec: %f\n", float64(totalUpdates) / duration.Seconds())
    log.Printf("%d\n", totalUpdates)

    /* Clean up */
    client.Cleanup()
}

func main() {
    /* Set up config */
    filename := flag.String("config", "src/config/client.config", "client config file")
    runTests := flag.Bool("test", false, "should run correctness tests")
    numDocs := flag.Int("num_docs", 0, "number of docs for artificial benchmark")
    benchmarkDir := flag.String("bench_dir", "", "directory containing files to benchmark")
    bloomFilterSz := flag.Int("bf_sz", 128, "bloom filter size in bits")
    isMalicious := flag.Bool("malicious", true, "run with malicious checks")
    fastSetup := flag.Bool("fast_setup", true, "run fast setup (ONLY TESTING)")
    useMaster := flag.Bool("use_master", true, "use a master for batched updates")
    runThroughput := flag.Bool("throughput", false, "run thorughput benchmarks")
    throughputSec := flag.Int("throughput_sec", 60, "throughput seconds")
    throughputThreads := flag.Int("throughput_threads", 64, "throughput threads")
    numUpdates := flag.Int("num_updates", 5, "number of updates before searches")
    numSearches := flag.Int("num_searches", 5, "number of searches before updates")
    numClusters := flag.Int("num_clusters", 0, "number of searches before updates")
    onlySetup := flag.Bool("only_setup", false, "only setup")
    flag.Parse()

    log.Println(*runTests)
    if (*runTests) {
        correctnessTests(*filename, *bloomFilterSz, *numDocs, *isMalicious, *useMaster)
    } else if (*onlySetup) {
        runFastSetup(*filename, *numDocs, *bloomFilterSz, *useMaster, *numClusters)
    }else if (*runThroughput && *numClusters == 0) {
        runThroughputBenchmark(*filename, *numDocs, *bloomFilterSz, *isMalicious, *useMaster, *throughputSec, *throughputThreads, *numUpdates, *numSearches)
    } else if (*runThroughput && *numClusters > 0) {
        runThroughputClustersBenchmark(*filename, *numDocs, *bloomFilterSz, *useMaster, *throughputSec, *throughputThreads, *numUpdates, *numSearches, *numClusters)

    } else if (*numClusters > 0) {
        runParallelBenchmark(*filename, *numDocs, *bloomFilterSz, *useMaster, *numClusters)
    }else if (*benchmarkDir != "") {
        runDirBenchmark(*filename, *benchmarkDir, *bloomFilterSz, *numDocs, *isMalicious, *useMaster)
    } else {
        runArtificialBenchmark(*filename, *numDocs, *bloomFilterSz, *isMalicious, *fastSetup, *useMaster)
    }
}
