import sys, string
from benchClient import runUpdateLatencyTest

bloomFilterSzList = [1120, 1280, 1440, 1600, 1800, 2000, 2240, 2520, 2800, 3120, 3480]

isMalicious = True

f = open("out/latency_update.dat", "w")

# Measure number of documents 2^10 to 2^20
for i in range(11):
    numDocs = 2 ** (i + 10)
    print(("Number of docs = %d, bloom filter size = %d") % (numDocs, bloomFilterSzList[i]))
    latency = runUpdateLatencyTest(bloomFilterSzList[i], numDocs, 10000, isMalicious)
    print("-------------------------")
    print(latency)
    f.write(str(latency) + "\n")

f.close()
