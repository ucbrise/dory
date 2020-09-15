import sys, string, json
import subprocess
import os
import threading
import time
import tempfile
import time

username = "ec2-user"

f_config = open('../system.config')
config = json.load(f_config)
f_config.close()
keyPath = config["SSHKeyPath"]
replicas = [server["Addr"] for server in config["Servers"]]
master = config["MasterAddr"]
clients = config["ClientAddrs"]
masterPort = config["MasterPort"]
replicaPorts = [server["Port"] for server in config["Servers"]]


devNull = open(os.devnull, 'w')

def generateRemoteCmdStr(machine, remoteCmd):
    return ("ssh -i %s -o StrictHostKeyChecking=no %s@%s \"%s\"") % (keyPath, username, machine, remoteCmd)

def generateDoryThroughputClientLocalStr(numDocs, bloomFilterSz, seconds, threads, isMalicious, numUpdates, numSearches, useMaster):
    return ("cd dory; ./runClient.sh -n %s -b %s -t true -x %s -y %s -m %s -q %s -r %s -s %s") % (numDocs, bloomFilterSz, seconds, threads, isMalicious, numUpdates, numSearches, useMaster)

def generateDoryMixedThroughputClientLocalStr(numDocs, bloomFilterSz, seconds, threads, numUpdates, numSearches, numClusters, isMalicious, isLeaky, useMaster):
    return ("cd dory; ./runClient.sh -n %s -b %s -t true -x %s -y %s -m %s -g %s -q %s -r %s -p %s -s %s") % (numDocs, bloomFilterSz, seconds, threads, isMalicious, isLeaky, numUpdates, numSearches, numClusters, useMaster)

def generateDorySetupClientLocalStr(numDocs, bloomFilterSz, numClusters, isMalicious):
    return ("cd dory; ./runClient.sh -n %s -b %s -p %s -z true -m %s") % (numDocs, bloomFilterSz, numClusters, isMalicious)

def generateDoryLatencyClientLocalStr(numDocs, bloomFilterSz, isMalicious, isLeaky, numClusters):
    return ("cd dory; ./runClient.sh -n %s -b %s -m %s -p %s -l true -a true -g %s") % (numDocs, bloomFilterSz, isMalicious, numClusters, isLeaky)

def generateUpdateLatencyClientLocalStr(numDocs, bloomFilterSz, isMalicious):
    return ("cd dory; ./runClient.sh -n %s -b %s -m %s -d ../maildir -u true") % (numDocs, bloomFilterSz, isMalicious)

def generateOramClientLocalStr(numDocs, oramServer):
    return("cd dory/baseline; export GOPATH=/home/ec2-user/dory/baseline; go run src/client/run_client.go -n %s -addr %s:4441") % (numDocs, oramServer)

def startDoryThroughputServers(bloomFilterSz, numDocs, tickMs):
    processes = []
    masterCmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null;cd dory; ./runMaster4.sh -n %s -b %s -t %s") % (masterPort, numDocs, bloomFilterSz, tickMs)
    print(masterCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(master, masterCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replica1Cmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %s -b %s -s 1") % (replicaPorts[0], numDocs, bloomFilterSz)
    print(replica1Cmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[0], replica1Cmd),
        shell=True, stdout=devNull, stderr=devNull))
    replica2Cmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %s -b %s -s 2") % (replicaPorts[1], numDocs, bloomFilterSz)
    print(replica2Cmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[1], replica2Cmd),
        shell=True, stdout=devNull, stderr=devNull))
    replica3Cmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %s -b %s -s 3") % (replicaPorts[2], numDocs, bloomFilterSz)
    print(replica1Cmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[2], replica1Cmd),
        shell=True, stdout=devNull, stderr=devNull))
    replica4Cmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %s -b %s -s 4") % (replicaPorts[3], numDocs, bloomFilterSz)
    print(replica2Cmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[3], replica2Cmd),
        shell=True, stdout=devNull, stderr=devNull))
    return processes

def startDoryLatencyServers(bloomFilterSz, numDocs, tickMs):
    processes = []
    masterCmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runMaster.sh -n %d -b %s -t %s") % (masterPort, int(numDocs), bloomFilterSz, tickMs)
    print(masterCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(master, masterCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replica1Cmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %s -b %s -s 1") % (replicaPorts[0], numDocs, bloomFilterSz)
    print(replica1Cmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[0], replica1Cmd),
        shell=True, stdout=devNull, stderr=devNull))
    replica2Cmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %s -b %s -s 2") % (replicaPorts[1], numDocs, bloomFilterSz)
    print(replica2Cmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[1], replica2Cmd),
        shell=True, stdout=devNull, stderr=devNull))
    return processes

def startParallelDoryLatencyServers(bloomFilterSz, numDocs, tickMs):
    processes = []
    masterCmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runMaster.sh -n %d -b %s -t %s") % (masterPort, numDocs, bloomFilterSz, tickMs)
    print(masterCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(master, masterCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %d -b %s -s 1") % (replicaPorts[0], numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[0], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %d -b %s -s 2") % (replicaPorts[1], numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[1], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %d -b %s -s 3") % (replicaPorts[2], numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[2], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %d -b %s -s 4") % (replicaPorts[3], numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[3], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %d -b %s -s 5") % (replicaPorts[4], numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[4], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %d -b %s -s 6") % (replicaPorts[5], numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[5], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %d -b %s -s 7") % (replicaPorts[6], numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[6], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    replicaCmd = ("sudo lsof -t -i tcp%s | sudo xargs kill > /dev/null &> /dev/null; cd dory; ./runServer.sh -n %d -b %s -s 8") % (replicaPorts[7], numDocs, bloomFilterSz)
    print(replicaCmd)
    processes.append(subprocess.Popen(generateRemoteCmdStr(replicas[7], replicaCmd),
        shell=True, stdout=devNull, stderr=devNull))
    return processes

def startOramServer(oramServer, numDocs):
    serverCmd = ("sudo lsof -t -i tcp:4441 | sudo xargs kill > /dev/null &> /dev/null; sleep 1; cd dory/baseline; export GOPATH=/home/ec2-user/dory/baseline; go run src/server/run_server.go -n %s") % (numDocs)
    print(serverCmd)
    process = subprocess.Popen(generateRemoteCmdStr(oramServer, serverCmd),
                shell=True, stdout=devNull, stderr=devNull)
    return process

def runSetupClient(clientLocalCmd):
    processes = []
    
    print("Starting setup client...")
    totalUpdates = 0
    print(clientLocalCmd)
    process = subprocess.Popen(generateRemoteCmdStr(clients[0], clientLocalCmd),
            shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    print("Started client")

    output = process.stderr.read()
    print(output)
    rawOutputLines = output.splitlines()
    outputLines = [line.decode('utf-8') for line in rawOutputLines]
    if len(outputLines) == 0:
        output = process.stdout.read()
    print("Done with setup.")


def runThroughputClients(clientLocalCmds, seconds):
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
        rawOutputLines = output.splitlines()
        outputLines = [line.decode('utf-8') for line in rawOutputLines]
        if len(outputLines) == 0:
            output = process.stdout.read()
            rawOutputLines = output.splitlines()
            outputLines = [line.decode('utf-8') for line in rawOutputLines]
        tokens =  outputLines[len(outputLines) -  1].split()
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

def runLatencyClient(clientLocalCmd):
    processes = []
    
    print("Starting all clients...")
    totalUpdates = 0
    print(generateRemoteCmdStr(clients[0], clientLocalCmd))
    process = subprocess.Popen(generateRemoteCmdStr(clients[0], clientLocalCmd),
            shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    print("Started client")

    output = process.stderr.read()
    print(output)
    rawOutputLines = output.splitlines()
    outputLines = [line.decode('utf-8') for line in rawOutputLines]
    if len(outputLines) == 0:
        output = process.stdout.read()
        print(output)
        rawOutputLines = output.splitlines()
        outputLines = [line.decode('utf-8') for line in rawOutputLines]
    tokens =  outputLines[len(outputLines) -  1].split()
    latency = float(tokens[len(tokens) - 1])

    print(("Latency = %s") % (tokens[len(tokens) - 1]))
    return tokens

def runOramClient(clientLocalCmd, oramClient):
    processes = []
    
    print("Starting ORAM client...")
    totalUpdates = 0
    print(clientLocalCmd)
    process = subprocess.Popen(generateRemoteCmdStr(oramClient, clientLocalCmd),
            shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    print("Started client")

    output = process.stderr.read()
    rawOutputLines = output.splitlines()
    outputLines = [line.decode('utf-8') for line in rawOutputLines]
    if len(outputLines) == 0:
        output = process.stdout.read()
        print(output.decode("utf-8"))

    return output.decode("utf-8") 

def runDoryThroughputTest(bloomFilterSz, numDocs, tickMs, clientS, threads, isMalicious, useMaster, numUpdates, numSearches):
    print (("Starting Dory with bloom filter size %s, num docs %s, tick ms %s, client duration %s, client threads %s, is malicious %s, updates to searches %s/%s") % (bloomFilterSz, numDocs, tickMs, clientS, threads, isMalicious, numUpdates, numSearches))

    servers = startDoryThroughputServers(bloomFilterSz, int(numDocs), tickMs)
    time.sleep(5)
    clientStrs = []
    for i in range(len(clients)):
        clientStrs.append(generateDoryThroughputClientLocalStr(numDocs, bloomFilterSz, clientS, threads, isMalicious, numUpdates, numSearches, useMaster))
    throughput = runThroughputClients(clientStrs, clientS)

    for i in range(len(servers)):
        servers[i].terminate()

    return throughput

def initForDoryMixedThroughput(bloomFilterSz, numDocs, tickMs, clientS, threads, numClusters, isMalicious):
    print (("Running setup with bloom filter size %s, num docs %s, tick ms %s, num clusters %s") % (bloomFilterSz, numDocs, tickMs, numClusters))

    servers = startParallelDoryLatencyServers(bloomFilterSz, int(numDocs), tickMs)
    time.sleep(5)
    runSetupClient(generateDorySetupClientLocalStr(numDocs, bloomFilterSz, numClusters, isMalicious))
    time.sleep(3)
    return servers 

def runDoryMixedThroughputTest(bloomFilterSz, numDocs, tickMs, clientS, threads, numUpdates, numSearches, numClusters, isMalicious, isLeaky, useMaster):
    print (("Starting Dory with bloom filter size %s, num docs %s, tick ms %s, client duration %s, client threads %s, updates to searches %s/%s, num clusters %s") % (bloomFilterSz, numDocs, tickMs, clientS, threads, numUpdates, numSearches, numClusters))

    clientStrs = []
    for i in range(len(clients)):
        clientStrs.append(generateDoryMixedThroughputClientLocalStr(numDocs, bloomFilterSz, clientS, threads, numUpdates, numSearches, numClusters, isMalicious, isLeaky, useMaster))
    throughput = runThroughputClients(clientStrs, clientS)
    return throughput

def cleanupForDoryMixedThroughput(servers):
    for i in range(len(servers)):
        servers[i].terminate()

def runUpdateLatencyTest(bloomFilterSz, numDocs, tickMs, isMalicious):
    print (("Starting Dory with bloom filter size %s, num docs %s, is malicious %s") % (bloomFilterSz, numDocs, isMalicious))

    servers = []
    servers = startDoryLatencyServers(bloomFilterSz, int(numDocs), tickMs)
    time.sleep(5)
    latencies = runLatencyClient(generateUpdateLatencyClientLocalStr(int(numDocs), bloomFilterSz, isMalicious))

    for i in range(len(servers)):
        servers[i].terminate()

    return latencies[len(latencies) - 1]

def runDoryLatencyTest(bloomFilterSz, numDocs, tickMs, isMalicious, isLeaky, numClusters):
    print (("Starting Dory with bloom filter size %s, num docs %s, is malicious %s, num clusters %s") % (bloomFilterSz, numDocs, isMalicious, numClusters))

    servers = []
    if numClusters == 0:
        servers = startDoryLatencyServers(bloomFilterSz, int(numDocs), tickMs)
    else: 
        servers = startParallelDoryLatencyServers(bloomFilterSz, int(numDocs), tickMs)
    time.sleep(5)
    latencies = runLatencyClient(generateDoryLatencyClientLocalStr(int(numDocs), bloomFilterSz, isMalicious, isLeaky, numClusters))

    for i in range(len(servers)):
        servers[i].terminate()

    return latencies

def runOramTest(oramServer, oramClient, numDocs):
    print("Starting ORAM tests for num docs %s" % numDocs)
    server = startOramServer(oramServer, numDocs)
    time.sleep(5)
    output = runOramClient(generateOramClientLocalStr(numDocs, oramServer), oramClient)
    server.terminate()
    return output
