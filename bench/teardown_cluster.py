import sys, string, json
import subprocess
import benchClient

filename = "../system.config"

f_config = open(filename)
sysConfig = json.load(f_config)

cmd = ('export AWS_DEFAULT_REGION=us-east-1; aws ec2 terminate-instances --instance-ids "%s"') % (sysConfig["MasterID"])
print(cmd)
process = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE)
process.wait()

'''
cmd = ('export AWS_DEFAULT_REGION=us-west-2; aws ec2 terminate-instances --instance-ids "%s"') % (sysConfig["BaselineServerID"])
print(cmd)
process = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE)
process.wait()
'''
cmd = ('export AWS_DEFAULT_REGION=us-east-1; aws ec2 terminate-instances --instance-ids "%s"') % (sysConfig["ClientIDs"][0])
print(cmd)
process = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE)
process.wait()
'''
cmd = ('export AWS_DEFAULT_REGION=us-west-1; aws ec2 terminate-instances --instance-ids "%s"') % (sysConfig["BaselineClientID"])
print(cmd)
process = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE)
process.wait()
'''
for i in range(len(sysConfig["Servers"])):
    region = "us-east-1"
    if i % 2 != 0:
        region = "us-east-2"
    print(cmd)
    cmd = ('export AWS_DEFAULT_REGION=%s; aws ec2 terminate-instances --instance-ids "%s"') % (region, sysConfig["Servers"][i]["ID"])
    process = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE)
    process.wait()

print("Finished cluster teardown")

'''
cmd = "rm ~/.ssh/dory.pem; export AWS_DEFAULT_REGION=us-east-1; aws ec2 delete-key-pair --key-name DoryKeyPair; export AWS_DEFAULT_REGION=us-east-2; aws ec2 delete-key-pair --key-name DoryKeyPair"
process = subprocess.Popen(cmd, shell=True)
process.wait()

cmd = "export AWS_DEFAULT_REGION=us-east-1; aws ec2 delete-security-group --group-name DoryGroup; export AWS_DEFAULT_REGION=us-east-2; aws ec2 delete-security-group --group-name DoryGroup"
process = subprocess.Popen(cmd, shell=True)
process.wait()
'''
