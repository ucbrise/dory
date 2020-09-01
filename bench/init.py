import sys, string, json, time
import subprocess
import benchClient

regionAMIs = { 
        "us-east-1": "ami-034fd6e677d79ebb5",
        "us-east-2": "ami-0f972c03390d700fb",
        "us-west-1": "ami-063ff9879f81ae1c1",
        "us-west-2": "ami-0fb4b87cd10b60c95",
        }   

def runSetup():
    
    cmd = "aws ec2 create-key-pair --key-name DoryKeyPair --query 'KeyMaterial' --output text > ~/.ssh/dory.pem; sudo chmod 400 ~/.ssh/dory.pem; ssh-keygen -y -f ~/.ssh/dory.pem > $HOME/.ssh/id_rsa_DoryKeyPair.pub;"
    process = subprocess.Popen(cmd, shell=True)
    process.wait()

    cmd = "aws ec2 import-key-pair --key-name DoryKeyPair --public-key-material fileb://$HOME/.ssh/id_rsa_DoryKeyPair.pub --region us-east-1 ; aws ec2 import-key-pair --key-name DoryKeyPair --public-key-material fileb://$HOME/.ssh/id_rsa_DoryKeyPair.pub --region us-east-2; aws ec2 import-key-pair --key-name DoryKeyPair --public-key-material fileb://$HOME/.ssh/id_rsa_DoryKeyPair.pub --region us-west-1 ; aws ec2 import-key-pair --key-name DoryKeyPair --public-key-material fileb://$HOME/.ssh/id_rsa_DoryKeyPair.pub --region us-west-2"
    process = subprocess.Popen(cmd, shell=True)
    process.wait()
    
    regions = ["us-east-1", "us-east-2", "us-west-1", "us-west-2"]

    for region in regions:
        cmd = ('export AWS_DEFAULT_REGION=%s; aws ec2 create-security-group --group-name DoryGroup --description "Dory security group"') % (region)
        process = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE)
        out = process.stdout.read()
        if len(out) == 0:
            continue
        print("output", out)
        secGroupConfig = json.loads(out)
        secGroupID = secGroupConfig["GroupId"]

        cmd = ("export AWS_DEFAULT_REGION=%s; aws ec2 authorize-security-group-ingress --group-name DoryGroup --protocol tcp --cidr 0.0.0.0/0 --port 0-65535") % (region)
        process = subprocess.Popen(cmd, shell=True)
        process.wait()

        cmd = ("export AWS_DEFAULT_REGION=%s; aws ec2 authorize-security-group-egress --group-id %s --ip-permissions IpProtocol=tcp,FromPort=0,ToPort=65535,IpRanges='[{CidrIp=0.0.0.0/0}]'") % (region, secGroupID)
        process = subprocess.Popen(cmd, shell=True)
        process.wait()


runSetup()

print("Completed AWS initialization")
