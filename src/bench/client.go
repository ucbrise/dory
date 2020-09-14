package main

// TODO: figure out openmp and add back -fopenmp

import (
    "bufio"
    "client"
    "common"
    "log"
    "flag"
    "os"
    "io/ioutil"
    "io"
    "time"
    "sync"
    "strconv"
    "math/rand"
)

/* Run correctness tests. */
func correctnessTests(configFile string, bloomFilterSz int, numDocs int, isMalicious bool, useMaster bool, inputDir string) {
    /* Initialize client */
    log.Println("Initializing client...")
    client.Setup(configFile, bloomFilterSz, numDocs)
    conn := client.OpenConnection()

    files,_ := ioutil.ReadDir(inputDir)
    numDocs = len(files)
    for docID, file := range files {
            var err error
            docID = docID % numDocs
            filename := inputDir + "/" + file.Name()
            if (isMalicious) {
                err = client.UpdateDocFile_malicious(conn, filename, docID, useMaster)
            } else {
                err = client.UpdateDocFile_semihonest(conn, filename, docID, useMaster)
            }
            if err != nil {
                log.Fatal(err)
            }
    }


    log.Println("Finished updates")

    time.Sleep(5000 * time.Millisecond)
    
    pass := true
    for docID, file := range files {
        keywords := client.GetKeywordsFromFile(inputDir + "/" + file.Name())
        for _, keyword := range keywords {
            var docs []byte
            if (isMalicious) {
                docs, _, _, _, _, _, _  = client.SearchKeyword_malicious(conn, keyword, useMaster)
            } else {
                docs, _ = client.SearchKeyword_semihonest(conn, keyword, useMaster)
            }
            if (docs[docID / 8] & (1 << (uint(docID) % 8)) == 0) {
                log.Printf("ERROR: did not find keyword %s in document %d\n", keyword, docID)
                pass = false
            }
        }
    }
    if (pass) {
        log.Printf("---- PASSED TESTS ----\n");
    } else {
        log.Printf("---- FAILED TESTS ----\n");
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

func runInteractiveSearches(configFile string, numDocs int, bloomFilterSz int, isMalicious bool, useMaster bool, inputDir string) {
    log.Println("Initializing client...")
    client.Setup(configFile, bloomFilterSz, numDocs)
    conn := client.OpenConnection()

    files,_ := ioutil.ReadDir(inputDir)
    numDocs = len(files)
    for docID, file := range files {
            var err error
            docID = docID % numDocs
            filename := inputDir + "/" + file.Name()
            if (isMalicious) {
                err = client.UpdateDocFile_malicious(conn, filename, docID, useMaster)
            } else {
                err = client.UpdateDocFile_semihonest(conn, filename, docID, useMaster)
            }
            if err != nil {
                log.Fatal(err)
            }
    }
    log.Println("Finished updates")

    input := bufio.NewScanner(os.Stdin)

    log.Printf("Enter a keyword to search for: ")

    for input.Scan() {
        var docs []byte
        keyword := client.StemWord(input.Text())
        if (isMalicious) {
            docs, _, _, _, _, _, _  = client.SearchKeyword_malicious(conn, keyword, useMaster)
        } else {
            docs, _ = client.SearchKeyword_semihonest(conn, keyword, useMaster)
        }
        log.Printf("Found keyword in: \n")
        found := false
        for i := uint(0); i < uint(numDocs); i++ {
            if (docs[i / 8] & (1 << (i % 8)) != 0) {
                log.Printf("... present in document %d\n", i)
                found = true
            }
        }
        if (!found) {
            log.Printf("... did not find keyword\n")
        }
        log.Printf("Enter a keyword to search for: ")
    }


    log.Println("Finished updates")
}

/* Run search latency benchmark without parallelism, including time breakdown. */
func runArtificialBenchmark(configFile string, numDocs int, bloomFilterSz int, isMalicious bool, isLeaky bool, isPlaintext bool, fastSetup bool, useMaster bool, latencyPrints bool) {
    /* Initialize client */

    numTrials := 1

    log.Println("Initializing client...")
    outFile := client.Setup(configFile, bloomFilterSz, numDocs)
    var conn *common.Conn

    if (useMaster && !isPlaintext) {
        conn = client.OpenConnection()
    }

    if (fastSetup) {
        client.RunFastSetup("", useMaster);
    } else {
        for i := 0; i < numDocs; i++ {
            if (isMalicious) {
                client.UpdateDoc_malicious(conn, []string{"hello", "world"}, i, useMaster)
            } else if (isPlaintext) {
                client.UpdateDoc_plaintext([]string{"hello", "world"}, i)
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
    clientMs := 0.0
    networkMs := 0.0
    serverMs := 0.0
    for i := 0; i < numTrials; i++ {
        if (isMalicious) {
            _, _, t1, t2, t3, t4, t5  := client.SearchKeyword_malicious(conn, "hello", useMaster)
            getStateMs += float64(t1.Nanoseconds())/float64(1e6)
            serverMs += float64(t4.Nanoseconds())/float64(1e6)
            clientMs += float64(t2.Nanoseconds())/float64(1e6) + float64(t5.Nanoseconds())/float64(1e6)
            networkMs = float64(t3.Nanoseconds())/float64(1e6) - serverMs
        } else if (isLeaky) {
            _, err = client.SearchKeyword_leaky(conn, "hello", useMaster)
        } else if (isPlaintext) {
            _, err = client.SearchKeyword_plaintext("hello")
        }else {
            _, err = client.SearchKeyword_semihonest(conn, "hello", useMaster)
        }
    }
    getStateMs = getStateMs/float64(numTrials)
    clientMs = clientMs/float64(numTrials)
    networkMs = networkMs/float64(numTrials)
    serverMs = serverMs/float64(numTrials)

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
    _, err = io.WriteString(file, strconv.FormatFloat(clientMs, 'f', 3, 64))
    _, err = io.WriteString(file, strconv.FormatFloat(networkMs, 'f', 3, 64))
    _, err = io.WriteString(file, strconv.FormatFloat(serverMs, 'f', 3, 64))
    if err != nil {
        log.Println(err)
    }

    if (latencyPrints) {
        log.Printf("total time to search: %f ms\n", timeMs);
        log.Printf("-> consensus: %f ms\n", getStateMs);
        log.Printf("-> client: %f ms\n", clientMs);
        log.Printf("-> network: %f ms\n", networkMs);
        log.Printf("-> server: %f ms\n", serverMs);
        log.Println("Cleaning up...")
        log.Printf("%f %f %f %f %f\n", getStateMs, clientMs, networkMs, serverMs, timeMs)
    } else {
        log.Printf("Completed search in %f ms\n", timeMs);
    }
    if (useMaster && !isPlaintext) {
        client.CloseConnection(conn)
    }
    client.Cleanup()
}

/* Measure update latency using documents from directory. */
func runDirBenchmark(configFile string, benchmarkDir string, bloomFilterSz int, numDocs int, isMalicious bool, isPlaintext bool, useMaster bool) {
    topDirs, err := ioutil.ReadDir(benchmarkDir)
    if err != nil {
        log.Fatal(err)
    }
    //totalIterations := 1000

    /* Initialize client */
    log.Println("Initializing client...")
    outFile := client.Setup(configFile, bloomFilterSz, numDocs)

    var conn *common.Conn

    if (useMaster && !isPlaintext) {
        conn = client.OpenConnection()
    }

    log.Println("Completed setup.")

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
                } else if (isPlaintext) {
                    err = client.UpdateDocFile_plaintext(filename, docID)
                }else {
                    err = client.UpdateDocFile_semihonest(conn, filename, docID, useMaster)
                }
                ctr += 1
                if err != nil {
                    log.Fatal(err)
                }
                /*if ctr >= totalIterations {
                    break
                }*/
            }
            /*if ctr >= totalIterations {
                break
            }*/
            log.Println("Finished: ", topDir.Name() + "/" + midDir.Name())
        }
        /*if ctr >= totalIterations {
            break
        }*/
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

    log.Printf("Average update time: %s ms\n", strconv.FormatFloat(elapsed, 'f', 3, 64))
    log.Printf("%s\n", strconv.FormatFloat(elapsed, 'f', 3, 64))

    /* Clean up */
    if (useMaster && !isPlaintext) {
        client.CloseConnection(conn)
    }
    client.Cleanup()
}

/* Start fast setup at servers for benchmarking/testing. */
func runFastSetup(configFile string, numDocs int, bloomFilterSz int, useMaster bool, numClusters int) {
    log.Println("Initializing client...")
    client.Setup(configFile, bloomFilterSz, numDocs)

    client.RunFastSetup_parallel("", useMaster, numClusters);
}

/* Run throughput benchmarks with  mix of updates and searches with multiple clusters. */
func runThroughputClustersBenchmark(configFile string, numDocs int, bloomFilterSz int, isMalicious bool, leaky bool, isPlaintext bool, useMaster bool, seconds int, threads int, numUpdates int, numSearches int, numClusters int) {
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
            var conn *common.Conn
            if (useMaster && !isPlaintext) {
                log.Println("opening conn to master")
                conn = client.OpenConnection()
                defer client.CloseConnection(conn)
            }
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
                default:
                    if (updateCtr < numUpdates) {
                        if (isMalicious) {
                            client.DummyUpdateDoc_malicious(conn, []string{"hello", "world"}, (j % numDocs), useMaster)
                        } else if (isPlaintext) {
                            docID := (j % 128) + 1
                            client.UpdateDocFile_plaintext("sample_docs/" + string(docID), docID)
                        } else {
                            client.DummyUpdateDoc_semihonest(conn, []string{"hello", "world"}, (j % numDocs), useMaster)
                        }
                        updateCtr += 1
                        if (updateCtr == numUpdates) {
                            searchCtr = 0
                        }
                    } else if (searchCtr < numSearches) {
                        if (isMalicious) {
                            log.Println("is malicious")
                            client.SearchKeyword_malicious_parallel(conn, "hello", useMaster, numClusters)
                        } else if (leaky) {
                            client.SearchKeyword_leaky(conn, "hello", useMaster)
                        } else if (isPlaintext) {
                            keywords := client.GetKeywordsFromFile("sample_docs/" + string((j % 128) + 1))
                            keyword := keywords[rand.Intn(len(keywords))]
                            client.SearchKeyword_plaintext(keyword)
                        } else {
                            client.SearchKeyword_semihonest(conn, "hello", useMaster)
                        }
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
        tag = "malicious_numUpdates"
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
    updateBench := flag.Bool("update_bench", false, "run update bnehcmarks")
    bloomFilterSz := flag.Int("bf_sz", 128, "bloom filter size in bits")
    isMalicious := flag.Bool("malicious", true, "run with malicious checks")
    leaky := flag.Bool("leaky", false, "run leaky version (no DPFs)")
    plaintext := flag.Bool("plaintext", false, "run plaintext version of search")
    fastSetup := flag.Bool("fast_setup", true, "run fast setup (ONLY TESTING)")
    useMaster := flag.Bool("use_master", true, "use a master for batched updates")
    runThroughput := flag.Bool("throughput", false, "run thorughput benchmarks")
    throughputSec := flag.Int("throughput_sec", 60, "throughput seconds")
    throughputThreads := flag.Int("throughput_threads", 64, "throughput threads")
    numUpdates := flag.Int("num_updates", 5, "number of updates before searches")
    numSearches := flag.Int("num_searches", 5, "number of searches before updates")
    numClusters := flag.Int("num_clusters", 0, "number of searches before updates")
    onlySetup := flag.Bool("only_setup", false, "only setup")
    latencyPrints := flag.Bool("latency_prints", false, "print out extra latency information")
    latencyBench := flag.Bool("latency_bench", false, "run latency benchmarks")
    flag.Parse()

    if (*runTests) {
        log.Println("going to run correctness tests")
        correctnessTests(*filename, *bloomFilterSz, *numDocs, *isMalicious, *useMaster, *benchmarkDir)
    } else if (*onlySetup) {
        runFastSetup(*filename, *numDocs, *bloomFilterSz, *useMaster, *numClusters)
    }else if (*runThroughput && *numClusters == 0) {
        runThroughputBenchmark(*filename, *numDocs, *bloomFilterSz, *isMalicious, *useMaster, *throughputSec, *throughputThreads, *numUpdates, *numSearches)
    } else if (*runThroughput && *numClusters > 0) {
        runThroughputClustersBenchmark(*filename, *numDocs, *bloomFilterSz, *isMalicious, *leaky, *plaintext, *useMaster, *throughputSec, *throughputThreads, *numUpdates, *numSearches, *numClusters)
    } else if (*updateBench) {
        runDirBenchmark(*filename, *benchmarkDir, *bloomFilterSz, *numDocs, *isMalicious, *plaintext, *useMaster)
    } else if (*latencyBench) {
        runArtificialBenchmark(*filename, *numDocs, *bloomFilterSz, *isMalicious, *leaky, *plaintext, *fastSetup, *useMaster, *latencyPrints)
    } else if (*numClusters > 0) {
        runParallelBenchmark(*filename, *numDocs, *bloomFilterSz, *useMaster, *numClusters)
    } else {
        runInteractiveSearches(*filename, *numDocs, *bloomFilterSz, *isMalicious, *useMaster, *benchmarkDir)
    }
}
