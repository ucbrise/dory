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

cmd = ("scp -i %s -o StrictHostKeyChecking=no %s@%s:dory/baseline/out/oram_1024 out/oram_1024") % (keyPath, username, baselineClientAddr)
process = subprocess.Popen(cmd, shell=True)
process.wait()

cmd = ("scp -i %s -o StrictHostKeyChecking=no %s@%s:dory/baseline/out/oram_2048 out/oram_2048") % (keyPath, username, baselineClientAddr)
process = subprocess.Popen(cmd, shell=True)
process.wait()

print("Results of baseline copied to out/oram_1024 and out/oram_2048")
