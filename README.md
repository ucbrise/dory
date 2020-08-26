# DORY: An Encrypted Search System with Distributed Trust

DORY is an encrypted search system that splits trust between multiple servers in order to efficiently hide access patterns from a malicious attacker who controls all but one of the servers. This implementation does NOT include the underlying end-to-end encrypted filesystem.

**WARNING**: This is an academic proof-of-concept prototype and has not received careful code review. This implementation is NOT ready for production use.

This prototype is released under the Apache v2 license (see [License](#license)).

## Setup

1. Run `git clone https://github.com/ucbrise/dory` locally. Make sure python, Golang, and matplotlib are downloaded.

2. Create the following EC2 instances (on-demand or spot) using the VM image provided:

| Instance type | Region | Quantity |
| --------------|:------:|:--------:|
| `r5n.4xlarge` | `east-1` | 5 |
| `c5.large` | `east-1` | 1 |
| `r5n.4xlarge` | `east-2` | 4 |
| `c5.large` | `west-1` | 1 |
| `r5n.4xlarge` | `west-2` | 1 |

This configuration was the one we used to generate our evaluation results, but you can also use different instsance types or regions (although you may obtain different results).

Label 1 `r5n.4xlarge east-1` instance `master`, and the other 4 `server1, server2, server3, server4`. Label the `c5.large` instance `client`. Label the 4 `r5n.4xlarge east-2` instances `server5, server6, server7, server8`. Label the 1 `c5.large west-1` instance `baseline-client` and the 1 `r5n.4xlarge west-2` instance `baseline-server`.

To use our configuration scripts, make sure that you can access all of the instances using the same SSH key.

Also, make sure to configure security groups so that each machine can be accessed via SSH (port 22) and each machine can contact each other. For simplicity, you can create one security group that is very permissive and each instance is a part of.

3. Open `system.config` locally. Update the `MasterAddr` field to be the IP address of `master`, `ClientAddr` to be an array containing the IP address of `client`, and the `Addr` field in the list of `Servers` to be the IP address of servers 1-8.

Set `SSHKeyPath` to be the path to the SSH key used to access all the instances.

Default TLS keys and certificates are included for testing. You do not need to change these to run evaluation benchmarks, but in a real deployment, these should be freshly generated for security.

4. Run `./setup.sh` locally. This will copy your local configuration to the EC2 instances you just created.

You've just finished setup! Follow the steps below to run experiments and reproduce our results.

**Note**: If you restart your instances and they are assigned different IP addresses, you will need to run steps 3 and 4 again.

## Running experiments

These scripts use SSH to run experiments. If your version of SSH is configured to check for known hosts (the default configuration), this will cause the scripts to hang unless you have SSH'd into each of the machines before. If you do not want to SSH into each machine individually before running the experiments, add the following line to `~/.ssh/config`:

```
Host *
    StrictHostKeyChecking no
```
 [maybe can take this out]

### Table 7

Run the experiment to collect the data for part of Table 7 showing the breakdown of search latency:

```
cd bench
python3 exp_tab7.py     # TODO: Add time estimate
```

This will produce data closely matching the left half of Table 7 on page 10 of the paper in `bench/out/tab7.dat`. For simplicity, we only show the numbers for one degree of parallelism (we exclude the two right-most columns). The affect of parallelism is shown in Figures 8b and 8c.

### Baseline

The experimental results in this paper compare DORY to a PathORAM baseline in `baseline/`. Unfortunately, running the experiments to produce the data in our paper takes about a week. The experiments will validate our baseline results for 1,024 and 2,048 documents. We will explain how to compare the results you produce to the ones in our paper, and then for the subsequent figures, we will use the results we produced for the baseline to reproduce the figures in our paper.

Because this experiment takes several hours to run and running locally can lead to broken SSH pipes in some situations, SSH into `baseline-client` directly. Run the following:
```
cd dory/baseline;
./runTests.sh       # TODO: Add time estimate
``` 

TODO: explain how to compare to output

### Figures 8b-8c

Run the experiment and then plot the data for Figures 8b and 8c showing the effect of parallelism on search latency as the number of documents increases:

```
cd bench
python3 exp_fig8b-c.py      # TODO: Add time estimate
python3 plot_fig8b.py       # few seconds
python3 plot_fig8c.py       # few seconds
```

This will produce plots close to Figures 8b and 8c on page 11 of the paper in `bench/out/fig8b.png` and `bench/out/fig8c.png`. Note that these plotting scripts use the data we collected for the baseline (in `bench/ref`) rather than experimental data, and we show how to validate the data we collected for the baseline at a reduced scale above.

### Figures 10-11

Run the experiment and then plot the data for Figures 10 and 11 showing the effect of parallelism on throughput as the number of documents increases for different workloads:

```
cd bench
python3 exp_fig10-11.py     # TODO: Add time estimate
python3 plot_fig10a.py      # few seconds
python3 plot_fig10b.py      # few seconds
python3 plot_fig10c.py      # few seconds
python3 plot_fig11a.py      # few seconds
python3 plot_fig11b.py      # few seconds
python3 plot_fig11c.py      # few seconds
```

This will produce plots close to Figures 10 and 11 on page 11 of the paper in `bench/out/fig10a.png`, `bench/out/fig10b.png`, `bench/out/fig10c.png`, `bench/out/fig11a.png`, `bench/out/fig11b.png`, `bench/out/fig11c.png`. Again, these plotting scripts use the data collected for the baseline (in `bench/ref`) rather than experimental data, and we show how to validate the data we collected for the baseline at a reduced scale above.

## Stand-alone usage
Start the master by running `runMaster.sh`, the servers by running `runServer.sh` and the client by running `runClient.sh`. Each script has a number of flags that can be set; run the scripts with `-h` to see all the flags.

For example, to start test DORY on a single machine (with two servers), use the default config files and run:
```
./runMaster.sh
./runServer.sh -s 1
./runServer.sh -s 2
./runClient.sh
```

Without any flags set, the client will load all the documents in `sample_docs` (a very smal subset of the Enron email dataset) into the search index and then provide prompts for the user to enter keywords to search for.

Make sure to always set the Bloom filter size and the max number of documents the same across the master, servers, and clients. The only exception is when running with cluster sizes greater than 1; in this case, every entity should use the same Bloom filter size, the master and client should use the correct maximum number of documents, and the servers should use the maximum  number of documents divided by the number of clusters. To run with multiple clusters, you will need a number of servers equal to 2 times the number of clusters.

## Tests
To test the low-level crypto, run `make` in `src/c` and run `correctness_tests`. To test the end-to-end system, run the client with the correctness test flag set to true (`-c`).

## Benchmarks
Use the scripts in `bench/` to run latency and throughput benchmarks on the system. Update the scripts with the IP addresses and ports of the different entities before running.

## Building from source

If installing from source instead, follow the below instructions:
1. Install OpenSSL, tested up to version 2.6.5.
2. Run `go get github.com/hashicorp/go-msgpack/codec`.
3. Download and build `libstemmer` (http://snowball.tartarus.org/download.html), tested up to version 2.0.0.
4. Move the output `libstemmer.o` to `src/c/`.



## Acknowledgements

We build on Saba Eskandarian's DPF implementation in https://github.com/SabaEskandarian/Express.
