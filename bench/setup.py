import sys, string, json
import subprocess
import benchClient

filename = "../system.config"

f_config = open(filename)
sysConfig = json.load(f_config)

sshKeyPath = sysConfig["SSHKeyPath"]

replicaPorts = [server["Port"] for server in sysConfig["Servers"]]
replicas = [server["Addr"] for server in sysConfig["Servers"]]

masterConfig = {
        "MasterAddr": sysConfig["MasterAddr"],
        "MasterPort": sysConfig["MasterPort"],
        "Addr": replicas,
        "Port": replicaPorts,
        "CertFile": sysConfig["MasterCertFile"],
        "KeyFile": sysConfig["MasterKeyFile"],
        "OutDir": sysConfig["OutDir"]
    }
masterConfigBlob = json.dumps(masterConfig)

with open("../src/config/master.config", "w") as f:
    f.write(masterConfigBlob)

if sysConfig["MasterAddr"] != "127.0.0.1":
    cmd = ("scp -i %s -o StrictHostKeyChecking=no $(PWD)/../src/config/master.config ec2-user@%s:~/dory/src/config/master.config") % (sshKeyPath, sysConfig["MasterAddr"])
    process = subprocess.Popen(cmd, shell=True)
    process.wait()

for i in range(len(sysConfig["Servers"])):
    serverConfig = {
            "Addr": sysConfig["Servers"][i]["Addr"],
            "Port": sysConfig["Servers"][i]["Port"],
            "CertFile": sysConfig["Servers"][i]["CertFile"],
            "KeyFile": sysConfig["Servers"][i]["KeyFile"],
            "OutDir": sysConfig["OutDir"],
            "ClientMaskKey": sysConfig["ClientMaskKey"],
            "ClientMacKey": sysConfig["ClientMacKey"]
        }
    serverNum = i + 1
    serverConfigBlob = json.dumps(serverConfig)
    
    with open(("../src/config/server%d.config") % (serverNum), "w") as f:
        f.write(serverConfigBlob)
    
    if sysConfig["Servers"][i]["Addr"] != "127.0.0.1":
        cmd = ("scp -i %s -o StrictHostKeyChecking=no $(PWD)/../src/config/server%d.config ec2-user@%s:~/dory/src/config/server%d.config") % (sshKeyPath, serverNum, sysConfig["Servers"][i]["Addr"], serverNum)
        process = subprocess.Popen(cmd, shell=True)
        process.wait()

clientConfig = {
        "MasterAddr": sysConfig["MasterAddr"],
        "MasterPort": sysConfig["MasterPort"],
        "Addr": replicas,
        "Port": replicaPorts,
        "MaskKey": sysConfig["ClientMaskKey"],
        "MacKey": sysConfig["ClientMacKey"]
    }
clientConfigBlob = json.dumps(clientConfig)
    
with open("../src/config/client.config", "w") as f:
    f.write(clientConfigBlob)
    
for i in range(len(sysConfig["ClientAddrs"])):
    
    if sysConfig["ClientAddrs"][i] != "127.0.0.1":
        cmd = ("scp -i %s -o StrictHostKeyChecking=no $(PWD)/../src/config/client.config ec2-user@%s:~/dory/src/config/client.config") % (sshKeyPath, sysConfig["ClientAddrs"][i])
        process = subprocess.Popen(cmd, shell=True)
        process.wait()


