import sys, string
from benchClient import runDoryThroughputTest

# FILL IN
clients = ["1.2.3.4"]
master = "5.6.7.8"
replicas = ["1.1.1.1", "2.2.2.2", "3.3.3.3", "4.4.4.4", "5.5.5.5", "6.6.6.6", "7.7.7.7", "8.8.8.8"]

if len(sys.argv) < 8:
    print("Required arguments: bloom filter size, num documents, tick ms, client s, is malicious, num updates, num searches")
    exit
bloomFilterSz = int(sys.argv[1])
numDocs = int(sys.argv[2])
tickMs = int(sys.argv[3])
clientS = int(sys.argv[4])
isMalicious = sys.argv[5]
numUpdates = int(sys.argv[6])
numSearches = int(sys.argv[7])

f = open("out/scale_throughput_dory_" + ("malicious" if isMalicious.lower() == "true" else "semihonest") + "_" + str(numUpdates) + "_" + str(numSearches), "w")

# Measure throughput for increasing number of clients
for i in range(len(clients)):
    threads = 10 
    numClients = (i + 1)
    print(("Number of threads = %d, number of clients = %d") % (threads, numClients))
    throughput = runDoryThroughputTest(master, replicas, clients[:numClients], bloomFilterSz, numDocs, tickMs, clientS, threads, isMalicious, numUpdates, numSearches)
    print("-------------------------")
    f.write(str(throughput) + "\n")

f.close()
