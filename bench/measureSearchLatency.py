import sys, string
from benchClient import runDoryLatencyTest

# FILL IN
clients = ["1.2.3.4"]
master = "5.6.7.8"
replicas = ["1.1.1.1", "2.2.2.2"]

bloomFilterSzList = [1120, 1280, 1440, 1600, 1800, 2000, 2240, 2520, 2800, 3120, 3480]

if len(sys.argv) < 3:
    print("Required arguments: bloom filter size, num documents, tick ms, client s, is malicious")
    exit
isMalicious = sys.argv[1]
breakdown = sys.argv[2].lower() == "true"

f = open("out/latency_dory_" + ("malicious" if isMalicious.lower() == "true" else "semihonest"), "w")

# Measure number of documents 2^10 to 2^20
for i in range(11):
    numDocs = 2 ** (i + 10)
    print(("Number of docs = %d, bloom filter size = %d") % (numDocs, bloomFilterSzList[i]))
    latencies = runDoryLatencyTest(master, replicas, clients[0], bloomFilterSzList[i], numDocs, 10000, isMalicious, 0)
    print("-------------------------")
    print(latencies[len(latencies) - 1])
    if not breakdown: 
        f.write(str(latencies[len(latencies) - 1]) + "\n")
    else:
        for latency in latencies: 
            f.write(str(latency) + " ")
        f.write("\n")

f.close()
