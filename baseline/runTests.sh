echo $1
echo $2
export GOPATH=/home/ec2-user/dory/baseline
ssh -i $2 ec2-user@$1 -o StrictHostKeyChecking=no "sudo lsof -t -i tcp:4441 | sudo xargs kill > /dev/null &> /dev/null; sleep 1; cd dory/baseline; export GOPATH=/home/ec2-user/dory/baseline; nohup go run src/server/run_server.go -n 1024 &" &
sleep 10 
go run src/client/run_client.go -n 1024 -addr $1:4441 > out/oram_1024
echo "Finished 1024"
ssh -i $2 ec2-user@$1 -o StrictHostKeyChecking=no "sudo lsof -t -i tcp:4441 | sudo xargs kill > /dev/null &> /dev/null; sleep 1; cd dory/baseline; export GOPATH=/home/ec2-user/dory/baseline; nohup go run src/server/run_server.go -n 2048 &" &
sleep 10
go run src/client/run_client.go -n 2048 -addr $1:4441 > out/oram_2048
echo "Finished 2048"
