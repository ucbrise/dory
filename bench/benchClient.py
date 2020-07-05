import sys, string
import subprocess
import os
import threading
import time
import tempfile
import time

# FILL IN
keyPath = ""
username = ""

devNull = open(os.devnull, 'w')

def generateRemoteCmdStr(machine, remoteCmd):
    return ("ssh -i %s %s@%s \"%s\"") % (keyPath, username, machine, remoteCmd)

def generateDoryThroughputClientLocalStr(numDocs, bloomFilterSz, seconds, threads, isMalicious, numUpdates, numSearches):
    return ("./runClient.sh -n %s -b %s -t true -x %s -y %s -m %s -q %s -r %s") % (numDocs, bloomFilterSz, seconds, threads, isMalicious, numUpdates, numSearches)

def generateDoryMixedThroughputClientLocalStr(numDocs, bloomFilterSz, seconds, threads, numUpdates, numSearches, numClusters):
    return ("./runClient.sh -n %s -b %s -t true -x %s -y %s -m true -q %s -r %s -p %s") % (numDocs, bloomFilterSz, seconds, threads, numUpdates, numSearches, numClusters)

def generateDorySetupClientLocalStr(numDocs, bloomFilterSz, numClusters):
    return ("./runClient.sh -n %s -b %s -p %s -z true") % (numDocs, bloomFilterSz, numClusters)

def generateDoryLatencyClientLocalStr(numDocs, bloomFilterSz, isMalicious, numClusters):
    return ("./runClient.sh -n %s -b %s -m %s -p %s") % (numDocs, bloomFilterSz, isMalicious, numClusters)

def generateUpdateLatencyClientLocalStr(numDocs, bloomFilterSz, isMalicious):
    return ("./runClient.sh -n %s -b %s -m %s -d ../maildir") % (numDocs, bloomFilterSz, isMalicious)

def startDoryThroughputServers(master, replicas, bloomFilterSz, numDocs, tickMs):
    processes = []
    masterCmd = ("./runMaster4.sh -n %s -b %s -t %s") % (numDocs, bloomFilterSz, tickMs)
    print(masterCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(master, masterCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replica1Cmd = ("./runServer.sh -n %s -b %s -s 1") % (numDocs, bloomFilterSz)
    print(replica1Cmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[0], replica1Cmd),
        shell=True, stdout=devNull, stderr=devNull))
    replica2Cmd = ("./runServer.sh -n %s -b %s -s 2") % (numDocs, bloomFilterSz)
    print(replica2Cmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[1], replica2Cmd),
        shell=True, stdout=devNull, stderr=devNull))
    replica3Cmd = ("./runServer.sh -n %s -b %s -s 3") % (numDocs, bloomFilterSz)
    print(replica1Cmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[2], replica1Cmd),
        shell=True, stdout=devNull, stderr=devNull))
    replica4Cmd = ("./runServer.sh -n %s -b %s -s 4") % (numDocs, bloomFilterSz)
    print(replica2Cmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[3], replica2Cmd),
        shell=True, stdout=devNull, stderr=devNull))
    return processes

def startDoryLatencyServers(master, replicas, bloomFilterSz, numDocs, tickMs):
    processes = []
    masterCmd = ("./runMaster.sh -n %d -b %s -t %s") % (int(numDocs), bloomFilterSz, tickMs)
    print(masterCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(master, masterCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replica1Cmd = ("./runServer.sh -n %s -b %s -s 1") % (numDocs, bloomFilterSz)
    print(replica1Cmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[0], replica1Cmd),
        shell=True, stdout=devNull, stderr=devNull))
    replica2Cmd = ("./runServer.sh -n %s -b %s -s 2") % (numDocs, bloomFilterSz)
    print(replica2Cmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[1], replica2Cmd),
        shell=True, stdout=devNull, stderr=devNull))
    return processes

def startParallelDoryLatencyServers(master, replicas, bloomFilterSz, numDocs, tickMs):
    processes = []
    masterCmd = ("./runMaster.sh -n %d -b %s -t %s") % (numDocs, bloomFilterSz, tickMs)
    print(masterCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(master, masterCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("./runServer.sh -n %d -b %s -s 1") % (numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[0], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("./runServer.sh -n %d -b %s -s 2") % (numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[1], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("./runServer.sh -n %d -b %s -s 3") % (numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[2], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("./runServer.sh -n %d -b %s -s 4") % (numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[3], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("./runServer.sh -n %d -b %s -s 5") % (numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[4], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("./runServer.sh -n %d -b %s -s 6") % (numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[5], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("./runServer.sh -n %d -b %s -s 7") % (numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[6], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("./runServer.sh -n %d -b %s -s 8") % (numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[7], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    return processes

def runSetupClient(client, clientLocalCmd):
    processes = []
    
    print("Starting setup client...")
    totalUpdates = 0
    print(clientLocalCmd)
    process = subprocess.Popen(generateRemoteCmdStr(client, clientLocalCmd),
            shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    print("Started client")

    output = process.stderr.read()
    print(output)
    outputLines = output.splitlines()
    if len(outputLines) == 0:
        output = process.stdout.read()
        print(output)
    print("Done with setup.")


def runThroughputClients(clients, clientLocalCmds, seconds):
    processes = []

    print(clientLocalCmds)
    print("Starting clients...")
    totalUpdates = 0
    processes = []
    for i in range(len(clients)):
        process = subprocess.Popen(generateRemoteCmdStr(clients[i], clientLocalCmds[i]),
                shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        processes.append(process)

    print("...started clients")

    start = time.time()
    for i in range(len(clients)):
        process = processes[i]
        while True:
            result = process.poll()
            if result is not None:
                break
            if time.time() >= start + (seconds * 6):
                print("FAIL: TIMEOUT")
                process.terminate()
                return -1
            time.sleep(1)

    for i in range(len(clients)):
        process = processes[i]
        output = process.stderr.read()
        outputLines = output.splitlines()
        if len(outputLines) == 0:
            output = process.stdout.read()
            outputLines = output.splitlines()
        tokens =  outputLines[len(outputLines) -  1].split()
        print(tokens)
        updates = -1
        try:
            updates = int(tokens[len(tokens) - 1])
        except:
            print("FAIL: didn't get back an int")
        print(("Updates from client %d = %d") % (i, updates))
        if totalUpdates >= 0:
            totalUpdates += updates

    print(("Total number of updates in %d seconds: %d") % (seconds, totalUpdates))
    print(f"Updates per sec: {totalUpdates / seconds:0.4f}")
    return totalUpdates / seconds

def runLatencyClient(client, clientLocalCmd):
    processes = []
    
    print("Starting all clients...")
    totalUpdates = 0
    print(clientLocalCmd)
    process = subprocess.Popen(generateRemoteCmdStr(client, clientLocalCmd),
            shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    print("Started client")

    output = process.stderr.read()
    print(output)
    outputLines = output.splitlines()
    if len(outputLines) == 0:
        output = process.stdout.read()
        print(output)
        outputLines = output.splitlines()
    tokens =  outputLines[len(outputLines) -  1].split()
    print(tokens)
    latency = float(tokens[len(tokens) - 1])

    print(("Latency = %s") % (tokens[len(tokens) - 1]))
    return tokens

def runOramClient(clientLocalCmd, client):
    processes = []
    
    print("Starting ORAM client...")
    totalUpdates = 0
    print(clientLocalCmd)
    process = subprocess.Popen(generateRemoteCmdStr(client, clientLocalCmd),
            shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    print("Started client")

    output = process.stderr.read()
    print(output)
    outputLines = output.splitlines()
    if len(outputLines) == 0:
        output = process.stdout.read()
        print(output.decode("utf-8"))

    return output.decode("utf-8") 

def runDoryThroughputTest(master, replicas, clients, bloomFilterSz, numDocs, tickMs, clientS, threads, isMalicious, numUpdates, numSearches):
    print (("Starting Dory with bloom filter size %s, num docs %s, tick ms %s, client duration %s, client threads %s, is malicious %s, updates to searches %s/%s") % (bloomFilterSz, numDocs, tickMs, clientS, threads, isMalicious, numUpdates, numSearches))

    servers = startDoryThroughputServers(master, replicas, bloomFilterSz, int(numDocs), tickMs)
    time.sleep(5)
    clientStrs = []
    for i in range(len(clients)):
        clientStrs.append(generateDoryThroughputClientLocalStr(numDocs, bloomFilterSz, clientS, threads, isMalicious, numUpdates, numSearches))
    throughput = runThroughputClients(clients, clientStrs, clientS)

    for i in range(len(servers)):
        servers[i].terminate()

    return throughput

def initForDoryMixedThroughput(master, replicas, clients, bloomFilterSz, numDocs, tickMs, clientS, threads, numClusters):
    print (("Running setup with bloom filter size %s, num docs %s, tick ms %s, num clusters %s") % (bloomFilterSz, numDocs, tickMs, numClusters))

    servers = startParallelDoryLatencyServers(master, replicas, bloomFilterSz, int(numDocs), tickMs)
    time.sleep(5)
    runSetupClient(clients[0], generateDorySetupClientLocalStr(numDocs, bloomFilterSz, numClusters))
    time.sleep(3)
    return servers 



def runDoryMixedThroughputTest(master, replicas, clients, bloomFilterSz, numDocs, tickMs, clientS, threads, numUpdates, numSearches, numClusters):
    print (("Starting Dory with bloom filter size %s, num docs %s, tick ms %s, client duration %s, client threads %s, updates to searches %s/%s, num clusters %s") % (bloomFilterSz, numDocs, tickMs, clientS, threads, numUpdates, numSearches, numClusters))

    clientStrs = []
    for i in range(len(clients)):
        clientStrs.append(generateDoryMixedThroughputClientLocalStr(numDocs, bloomFilterSz, clientS, threads, numUpdates, numSearches, numClusters))
    throughput = runThroughputClients(clients, clientStrs, clientS)

def cleanupForDoryMixedThroughput(servers):
    for i in range(len(servers)):
        servers[i].terminate()

def runUpdateLatencyTest(master, replicas, client, bloomFilterSz, numDocs, tickMs, isMalicious):
    print (("Starting Dory with bloom filter size %s, num docs %s, is malicious %s") % (bloomFilterSz, numDocs, isMalicious))

    servers = []
    servers = startDoryLatencyServers(master, replicas, bloomFilterSz, int(numDocs), tickMs)
    time.sleep(5)
    latencies = runLatencyClient(client, generateUpdateLatencyClientLocalStr(int(numDocs), bloomFilterSz, isMalicious))

    for i in range(len(servers)):
        servers[i].terminate()

    return latencies[len(latencies) - 1]

def runDoryLatencyTest(master, replicas, client, bloomFilterSz, numDocs, tickMs, isMalicious, numClusters):
    print (("Starting Dory with bloom filter size %s, num docs %s, is malicious %s, num clusters %s") % (bloomFilterSz, numDocs, isMalicious, numClusters))

    servers = []
    if numClusters == 0:
        servers = startDoryLatencyServers(master, replicas, bloomFilterSz, int(numDocs), tickMs)
    else: 
        servers = startParallelDoryLatencyServers(master, replicas, bloomFilterSz, int(numDocs), tickMs)
    time.sleep(5)
    latencies = runLatencyClient(client, generateDoryLatencyClientLocalStr(int(numDocs), bloomFilterSz, isMalicious, numClusters))

    for i in range(len(servers)):
        servers[i].terminate()

    return latencies

