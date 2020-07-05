import sys, string
from benchClient import runDoryMixedThroughputTest
from benchClient import initForDoryMixedThroughput
from benchClient import cleanupForDoryMixedThroughput

# FILL IN
clients = ["1.2.3.4"]
master = "5.6.7.8"
replicas = ["1.1.1.1", "2.2.2.2", "3.3.3.3", "4.4.4.4", "5.5.5.5", "6.6.6.6", "7.7.7.7", "8.8.8.8"]

bloomFilterSzList = [1120, 1280, 1440, 1600, 1800, 2000, 2240, 2520, 2800, 3120, 3480]

if len(sys.argv) < 8:
    print("Required arguments: bloom filter size, num documents, tick ms, client s, is malicious, num updates, num searches")
    exit
tickMs = int(sys.argv[1])
clientS = int(sys.argv[2])
numClusters = int(sys.argv[3])

threads = 10
numClients = 1


f91 = open("out/scale_mixed_throughput_dory_" + str(numClusters) + "_9_1", "w")
f55 = open("out/scale_mixed_throughput_dory_" + str(numClusters) + "_5_5", "w")
f19 = open("out/scale_mixed_throughput_dory_" + str(numClusters) + "_1_9", "w")

# Measure throughput for documents 2^10 to 2^20 for
# - 90% updates / 10% searches
# - 50% updates / 50% searches
# - 10% updates / 90% searches
for i in range(11):
    rawNumDocs = 2 ** (i + 10)

    numDocs = int(rawNumDocs / numClusters)
    bloomFilterSz = bloomFilterSzList[i]

    servers = initForDoryMixedThroughput(master, replicas, clients[:numClients], bloomFilterSz, numDocs, tickMs, clientS, threads, numClusters)

    numUpdates = 9
    numSearches = 1
    print(("Number of threads = %d, number of clients = %d") % (threads, numClients))
    throughput = runDoryMixedThroughputTest(master, replicas, clients[:numClients], bloomFilterSz, numDocs, tickMs, clientS, threads, numUpdates, numSearches, numClusters)
    print("-------------------------")
    f91.write(str(throughput) + "\n")

    numUpdates =5 
    numSearches = 5
    print(("Number of threads = %d, number of clients = %d") % (threads, numClients))
    throughput = runDoryMixedThroughputTest(master, replicas, clients[:numClients], bloomFilterSz, numDocs, tickMs, clientS, threads, numUpdates, numSearches, numClusters)
    print("-------------------------")
    f55.write(str(throughput) + "\n")

    numUpdates = 1
    numSearches = 9
    f19 = open("out/scale_mixed_throughput_dory_" + str(numClusters) + "_" + str(numUpdates) + "_" + str(numSearches), "w")
    print(("Number of threads = %d, number of clients = %d") % (threads, numClients))
    throughput = runDoryMixedThroughputTest(master, replicas, clients[:numClients], bloomFilterSz, numDocs, tickMs, clientS, threads, numUpdates, numSearches, numClusters)
    print("-------------------------")
    f19.write(str(throughput) + "\n")

    cleanupForDoryMixedThroughput(servers)

f91.close()
f55.close()
f19.close()
