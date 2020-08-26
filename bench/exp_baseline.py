import sys, string
from benchClient import runOramTest

if len(sys.argv) < 3:
    print("Need 2 arguments: client IP address, server IP address")
    exit

client = sys.argv[1]
server = sys.argv[2]

print(("Client IP address = %s") % (client))
print(("Server IP address = %s") % (server))


for i in range(2):
    numDocs = 2 ** (i + 10)
    print(("Number of docs = %d") % (numDocs))
    output = runOramTest(server, client, numDocs)
    print("-------------------------")
    print(output)
    f = open("out/oram_" + str(numDocs) , "w")
    f.write(str(output))
    f.close()
