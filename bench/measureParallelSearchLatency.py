import sys, string
from benchClient import runDoryLatencyTest

# FILL IN
clients = ["1.2.3.4"]
master = "5.6.7.8"
replicas = ["1.1.1.1", "2.2.2.2", "3.3.3.3", "4.4.4.4", "5.5.5.5", "6.6.6.6", "7.7.7.7", "8.8.8.8"]

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

if len(sys.argv) < 2:
    print("Required arguments: bloom filter size, num documents, tick ms, client s, is malicious")
    exit
numDocs = int(sys.argv[1])

if numDocs not in bloomFilterSzDict:
    print("Bad num docs, no corresponding bloom filter size")
    exit

bloomFilterSz = bloomFilterSzDict[numDocs]
print(("Using bloom filter size %d for num docs %d") % (bloomFilterSz, numDocs))

f = open("out/latency_dory_parallel_" + str(numDocs), "w")

# Measure for number of clusters 1, 2, 4
for i in range(3):
    numClusters = 2 ** i
    currNumDocs = numDocs / numClusters 
    print(("Num clusters = %d, each clusters has %d docs") % (numClusters, currNumDocs))
    latencies = runDoryLatencyTest(master, replicas, clients[0], bloomFilterSz, currNumDocs, 10000, True, numClusters)
    print("-------------------------")
    print(latencies[len(latencies) - 1])
    f.write(str(latencies[len(latencies) - 1]) + "\n")

f.close()
