import sys, string, json, time
import subprocess
import benchClient

filename = "../local.config"

print("Generating local configuration files...")

f_config = open(filename, "r")
sysConfig = json.load(f_config)
f_config.close()

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

print("Finished generating local configuration files. Ready to run locally.")
