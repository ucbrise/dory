import sys, string
from benchClient import runDoryMixedThroughputTest
from benchClient import initForDoryMixedThroughput
from benchClient import cleanupForDoryMixedThroughput

bloomFilterSzList = [1120, 1280, 1440, 1600, 1800, 2000, 2240, 2520, 2800, 3120, 3480]

tickMs = 10000 
clientS = 10 
#clientS = 60 

threads = 10

def runExperimentRound(numClusters):

    f91 = open("out/dory_throughput_" + str(numClusters) + "_9_1.dat", "w")
    f55 = open("out/dory_throughput_" + str(numClusters) + "_5_5.dat", "w")
    f19 = open("out/dory_throughput_" + str(numClusters) + "_1_9.dat", "w")

    # Measure throughput for documents 2^10 to 2^20 for
    # - 90% updates / 10% searches
    # - 50% updates / 50% searches
    # - 10% updates / 90% searches
    for i in range(11):
        rawNumDocs = 2 ** (i + 10)

        numDocs = int(rawNumDocs / numClusters)
        bloomFilterSz = bloomFilterSzList[i]

        servers = initForDoryMixedThroughput(bloomFilterSz, numDocs, tickMs, clientS, threads, numClusters)

        numUpdates = 9
        numSearches = 1
        print(("9/1 for %d clusters, %d docs") % (numClusters, numDocs))
        throughput = runDoryMixedThroughputTest(bloomFilterSz, numDocs, tickMs, clientS, threads, numUpdates, numSearches, numClusters)
        print("-------------------------")
        f91.write(str(throughput) + "\n")

        numUpdates =5 
        numSearches = 5
        print(("5/5 for %d clusters, %d docs") % (numClusters, numDocs))
        throughput = runDoryMixedThroughputTest(bloomFilterSz, numDocs, tickMs, clientS, threads, numUpdates, numSearches, numClusters)
        print("-------------------------")
        f55.write(str(throughput) + "\n")

        numUpdates = 1
        numSearches = 9
        print(("1/9 for %d clusters, %d docs") % (numClusters, numDocs))
        throughput = runDoryMixedThroughputTest(bloomFilterSz, numDocs, tickMs, clientS, threads, numUpdates, numSearches, numClusters)
        print("-------------------------")
        f19.write(str(throughput) + "\n")

        cleanupForDoryMixedThroughput(servers)

    f91.close()
    f55.close()
    f19.close()

runExperimentRound(1)
runExperimentRound(2)
runExperimentRound(4)
