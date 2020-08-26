package main

import(
    "common"
    "encoding/json"
    "flag"
    "os"
    "os/exec"
    "fmt"
    "io/ioutil"
)

func main() {
    filename := flag.String("system_config", "system.config", "master system config")
    
    flag.Parse()
    
    sysConfig := common.SystemConfig{}
    file, err := os.Open(*filename)
    if err != nil {
        fmt.Errorf("Cannot open file: ", *filename)
    }
    defer file.Close()
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&sysConfig)
    if err != nil {
        fmt.Errorf("Cannot parse system config: ", *filename)
    }
    sshKeyPath := sysConfig.SSHKeyPath

    fmt.Println(sysConfig)

    addrs := make([]string, 0)
    ports := make([]string, 0)

    for i := 0; i < len(sysConfig.Servers); i += 1 {
        addrs = append(addrs, sysConfig.Servers[i].Addr)
        ports = append(ports, sysConfig.Servers[i].Port)
    }

    masterConfig := common.MasterConfig{
        MasterAddr: sysConfig.MasterAddr,
        MasterPort: sysConfig.MasterPort,
        Addr: addrs,
        Port: ports,
        CertFile: sysConfig.MasterCertFile,
        KeyFile: sysConfig.MasterKeyFile,
        OutDir: sysConfig.OutDir,
    }
    masterFile,_ := json.Marshal(masterConfig)
    err = ioutil.WriteFile("src/config/master.config", masterFile, 0644)

    fmt.Println("wrote master config")

    if (sysConfig.MasterAddr != "127.0.0.1") {
        dest := fmt.Sprintf("ec2-user@%s:~/dory/src/config/master.config", sysConfig.MasterAddr)
        fmt.Println(sshKeyPath)
        cmd := exec.Command("bash")
        cmdWriter,_ := cmd.StdinPipe()
        cmd.Start()
        cmdWriter.Write([]byte("scp -i " + sshKeyPath + " $(PWD)/src/config/master.config " + dest + "\n"))
        cmdWriter.Write([]byte("exit\n"))
        cmd.Wait()
        fmt.Println("finished scp to master")
    }

    for i := 0; i < len(sysConfig.Servers); i += 1 {
        serverConfig := common.ServerConfig{
            Addr: sysConfig.Servers[i].Addr,
            Port: sysConfig.Servers[i].Port,
            CertFile: sysConfig.Servers[i].CertFile,
            KeyFile: sysConfig.Servers[i].KeyFile,
            OutDir: sysConfig.OutDir,
            ClientMaskKey: sysConfig.ClientMaskKey,
            ClientMacKey: sysConfig.ClientMacKey,
        }
        serverNum := i + 1
        serverFileName := fmt.Sprintf("src/config/server%d.config", serverNum)
        serverFile,_ := json.Marshal(serverConfig)
        err = ioutil.WriteFile(serverFileName, serverFile, 0644)
        if (sysConfig.Servers[i].Addr != "127.0.0.1") {
            dest := fmt.Sprintf("ec2-user@%s:~/dory/src/config/server%d.config", sysConfig.Servers[i].Addr, serverNum)
            src := fmt.Sprintf("$(PWD)/src/config/server%d.config", serverNum)
            cmd := exec.Command("bash")
            cmdWriter,_ := cmd.StdinPipe()
            cmd.Start()
            cmdWriter.Write([]byte("scp -i " + sshKeyPath + " " + src + " " + dest + "\n"))
            cmdWriter.Write([]byte("exit\n"))
            cmd.Wait()
        }


    }

    clientConfig := common.ClientConfig{
        MasterAddr: sysConfig.MasterAddr,
        MasterPort: sysConfig.MasterPort,
        Addr: addrs,
        Port: ports,
        MaskKey: sysConfig.ClientMaskKey,
        MacKey: sysConfig.ClientMacKey,
    }
    clientFile,_ := json.Marshal(clientConfig)
    err = ioutil.WriteFile("src/config/client.config", clientFile, 0644)

    for i := 0; i < len(sysConfig.ClientAddrs); i += 1 {
        if (sysConfig.ClientAddrs[i] != "127.0.0.1") {
            dest := fmt.Sprintf("ec2-user@%s:~/dory/src/config/client.config", sysConfig.ClientAddrs[i])
            cmd := exec.Command("bash")
            cmdWriter,_ := cmd.StdinPipe()
            cmd.Start()
            cmdWriter.Write([]byte("scp -i " + sshKeyPath + " $(PWD)/src/config/client.config " + dest + "\n"))
            cmdWriter.Write([]byte("exit\n"))
            cmd.Wait()
        }
    }

}
