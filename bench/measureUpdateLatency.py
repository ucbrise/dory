import sys, string
from benchClient import runUpdateLatencyTest

# FILL IN
clients = ["1.2.3.4"]
master = "5.6.7.8"
replicas = ["1.1.1.1", "2.2.2.2"]

bloomFilterSzList = [1120, 1280, 1440, 1600, 1800, 2000, 2240, 2520, 2800, 3120, 3480]

if len(sys.argv) < 3:
    print("Required arguments: bloom filter size, num documents, tick ms, client s, is malicious")
    exit
isMalicious = sys.argv[1]

f = open("out/latency_update_" + ("malicious" if isMalicious.lower() == "true" else "semihonest"), "w")

# Measure number of documents 2^10 to 2^20
for i in range(11):
    numDocs = 2 ** (i + 10)
    print(("Number of docs = %d, bloom filter size = %d") % (numDocs, bloomFilterSzList[i]))
    latency = runUpdateLatencyTest(master, replicas, clients[0], bloomFilterSzList[i], numDocs, 10000, isMalicious)
    print("-------------------------")
    print(latency)
    f.write(str(latency) + "\n")

f.close()
