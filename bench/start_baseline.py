import sys, string, json, os
from benchClient import generateRemoteCmdStr
import subprocess

username = "ec2-user"

f_config = open('../system.config')
config = json.load(f_config)
f_config.close()
keyPath = config["SSHKeyPath"]
baselineServerAddr = config["BaselineServerAddr"]
baselineClientAddr = config["BaselineClientAddr"]
devNull = open(os.devnull, 'w')

cmd = ("scp -i %s -o StrictHostKeyChecking=no %s %s@%s:%s") % (keyPath, keyPath, username, baselineClientAddr, keyPath)
process = subprocess.Popen(cmd, shell=True)
process.wait()

cmd = ("cd dory/baseline; nohup ./runTests.sh %s %s &") % (baselineServerAddr, keyPath)
process = subprocess.Popen(generateRemoteCmdStr(baselineClientAddr, cmd) + " &", shell=True, stdout=devNull)
process.wait()

print(("Started baseline tests. In about 70 minutes, check dory/baseline/out/oram_1024 and dory/baseline/out/oram_2048 at IP address %s") % (baselineClientAddr))
