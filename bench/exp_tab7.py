import sys, string, json
from benchClient import runDoryLatencyTest

# FILL IN
clients = ["1.2.3.4"]
master = "5.6.7.8"
replicas = ["1.1.1.1", "2.2.2.2"]

bloomFilterSzList = [1120, 1280, 1440, 1600, 1800, 2000, 2240, 2520, 2800, 3120, 3480]

isMalicious = True 
isLeaky = False
breakdown = True

f = open("out/tab7.dat", "w")

# Measure number of documents 2^10 to 2^20
for i in range(11):
    numDocs = 2 ** (i + 10)
    print(("Number of docs = %d, Bloom filter size = %d") % (numDocs, bloomFilterSzList[i]))
    latencies = runDoryLatencyTest(bloomFilterSzList[i], numDocs, 10000, isMalicious, isLeaky, 0)
    print("-------------------------")
    f.write(("Number of docs = 2^%d, Bloom filter size = %d\n") % (i + 10, bloomFilterSzList[i]))
    if not breakdown: 
        print(("Time: %s ms") % latencies[2])
        f.write(str(latencies[len(latencies) - 1]) + "\n")
    else:
        print(("Total time: %s ms") % (latencies[6]))
        print(("-> Consensus time: %s ms") % (latencies[2]))
        print(("-> Client time: %s ms") % (latencies[3]))
        print(("-> Network time: %s ms") % (latencies[4]))
        print(("-> Server time: %s ms") % (latencies[5]))
        
        f.write(("Total time: %s ms\n") % (latencies[6]))
        f.write(("-> Consensus time: %s ms\n") % (latencies[2]))
        f.write(("-> Client time: %s ms\n") % (latencies[3]))
        f.write(("-> Network time: %s ms\n") % (latencies[4]))
        f.write(("-> Server time: %s ms\n\n") % (latencies[5]))
 
f.close()
