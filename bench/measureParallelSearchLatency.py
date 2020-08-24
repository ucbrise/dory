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

if len(sys.argv) < 2:
    print("Required arguments: bloom filter size, num documents, tick ms, client s, is malicious")
    exit
numDocs = int(sys.argv[1])

if numDocs not in bloomFilterSzDict:
    print("Bad num docs, no corresponding bloom filter size")
    exit

bloomFilterSz = bloomFilterSzDict[numDocs]
print(("Using bloom filter size %d for num docs %d") % (bloomFilterSz, numDocs))

f_config = open('../system.config')
config = json.load(f_config)
f_config.close()

replicas = [server["Addr"] for server in config["Servers"]]

f = open("out/latency_dory_parallel_" + str(numDocs), "w")

# Measure for number of clusters 1, 2, 4
for i in range(3):
    numClusters = 2 ** i
    currNumDocs = numDocs / numClusters 
    print(("Num clusters = %d, each clusters has %d docs") % (numClusters, currNumDocs))
    latencies = runDoryLatencyTest(config["MasterAddr"], replicas, config["ClientAddr"], bloomFilterSz, currNumDocs, 10000, True, numClusters)
    print("-------------------------")
    print(latencies[len(latencies) - 1])
    f.write(str(latencies[len(latencies) - 1]) + "\n")

f.close()
