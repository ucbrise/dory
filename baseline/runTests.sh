source ~/.bashrc
export GOPATH=~/dory/baseline
ssh -i ~/.ssh/oblivisearch-server.pem ec2-user@54.245.166.95 "sudo lsof -t -i tcp:4441 | sudo xargs kill > /dev/null &> /dev/null; sleep 1; cd dory/baseline; export GOPATH=~/dory/baseline; nohup go run src/server/run_server.go -n 1024 &" &
sleep 3
go run src/client/run_client.go -n 1024 > out/oram_1024
echo "Finished 1024"
ssh -i ~/.ssh/oblivisearch-server.pem ec2-user@54.245.166.95 "sudo lsof -t -i tcp:4441 | sudo xargs kill > /dev/null &> /dev/null; sleep 1; cd dory/baseline; export GOPATH=~/dory/baseline; nohup go run src/server/run_server.go -n 2048 &" &
sleep 3
go run src/client/run_client.go -n 2048 > out/oram_2048
echo "Finished 2048"

