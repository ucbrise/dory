import sys, string, json
from benchClient import runDoryLatencyTest

bloomFilterSzDict = {
        1024: 1120,
        2048: 1280,
        4096: 1440,
        8192: 1600,
        16384: 1800,
        32768: 2000,
        65536: 2240,
        131072: 2520,
        262144: 2800,
        524288: 3120,
        517408: 3120,
        1048576: 3480
}

f = []
f.append(open("out/latency_search_dory_1.dat", "w"))
f.append(open("out/latency_search_dory_2.dat", "w"))
f.append(open("out/latency_search_dory_4.dat", "w"))

# Measure for number of clusters 1, 2, 4
for i in range(11):
    numDocs = 2 ** (i + 10) 
    bloomFilterSz = bloomFilterSzDict[numDocs]
    print(("Using bloom filter size %d for num docs %d") % (bloomFilterSz, numDocs))
    for j in range(3):
        numClusters = 2 ** j
        currNumDocs = numDocs / numClusters 
        print(("Num clusters = %d, each clusters has %d docs") % (numClusters, currNumDocs))
        latencies = runDoryLatencyTest(bloomFilterSz, currNumDocs, 10000, True, numClusters)
        print("-------------------------")
        print(latencies[len(latencies) - 1])
        f[j].write(str(latencies[len(latencies) - 1]) + "\n")

f[0].close()
f[1].close()
f[2].close()
