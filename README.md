# DORY: An Encrypted Search System with Distributed Trust

DORY is an encrypted search system that splits trust between multiple servers in order to efficiently hide access patterns from a malicious attacker who controls all but one of the servers. This implementation does NOT include the underlying end-to-end encrypted filesystem.

**WARNING**: This is an academic proof-of-concept prototype and has not received careful code review. This implementation is NOT ready for production use.

This prototype is released under the Apache v2 license (see [License](#license)).

## Setup

1. Run `git clone https://github.com/ucbrise/dory` locally. Make sure python3 is downloaded. In `bench/`, run `pip3 install requirements.txt`.

2. Create the following EC2 instances (on-demand or spot) using the VM image provided:

| Instance type | Region | Quantity | Name(s) |
| --------------|:------:|:--------:|:-------:|
| `r5n.4xlarge` | `east-1` | 5 | `server1`, `server3`, `server5`, `server7`, `master` |
| `c5.large` | `east-1` | 1 | `client` |
| `r5n.4xlarge` | `east-2` | 4 | `server2`, `server4`, `server6`, `server8` |
| `c5.large` | `west-1` | 1 | `baseline-client` |
| `r5n.4xlarge` | `west-2` | 1 | `baseline-server` |

This configuration was the one we used to generate our evaluation results, but you can also use different instsance types or regions (although you may obtain different results).

To use our configuration scripts, make sure that you can access all of the instances using the same SSH key.

Also, make sure to configure security groups so that each machine can be accessed via SSH (port 22) and each machine can contact each other. For simplicity, you can create one security group that is very permissive and each instance is a part of.

3. Open `system.config` locally. Update the `MasterAddr` field to be the IP address of `master`, `ClientAddr` to be an array containing the IP address of `client`, and the `Addr` field in the list of `Servers` to be the IP address of servers 1-8.

Set `SSHKeyPath` to be the path to the SSH key used to access all the instances.

Default TLS keys and certificates are included for testing. You do not need to change these to run evaluation benchmarks, but in a real deployment, these should be freshly generated for security.

4. In `bench/`, run `python3 setup.py` locally. This will copy your local configuration to the EC2 instances you just created.

You've just finished setup! Follow the steps below to run experiments and reproduce our results.

**Note**: If you restart your instances and they are assigned different IP addresses, you will need to run steps 3 and 4 again.

## Running experiments

TODO: explain fast setup somewhere

The experimental results in this paper compare DORY to a PathORAM baseline in `baseline/`. Unfortunately, running the experiments to produce the data in our paper takes about a week. We will show how to validate our baseline results for a small number of documents and then for the other figures, we will use the results we produced for the baseline in order to reproduce the figures in our paper.

The experiments for Table 7, Figures 8b-8c, and Figures 10-11 cannot be run concurrently. However, the experiments for the baseline can be run at the same time as the DORY experiments (we recommend doing this to save time, as the baseline experiments take a few hours to complete).

To speed up testing, some of the experiments start with an index that is built by the server where the server has the keys to generate a correct search index. This configuration should only be used for testing (for security, only the client should have the keys).

### Table 7

Run the experiment to collect the data for part of Table 7 showing the breakdown of search latency. For this experiment, you only need to have `server-1`, `server-2`, `master`, and `client` running. Run the following commands locally:

```
cd bench
python3 exp_tab7.py     # 9 minutes 
```

This will produce data closely matching the left half of Table 7 on page 10 of the paper in `bench/out/tab7.dat`. You might notice some variation compared to the numbers displayed in the table for network latency. This is due to the fact that to quickly reproduce the results for this table, we are not averaging over many trials. Also, for simplicity, we only show the numbers for one degree of parallelism (we exclude the two right-most columns). The effect of parallelism is shown in Figures 8b and 8c.

<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/tab7.png" width="400">

### Figures 8b-8c

Run the experiment and then plot the data for Figures 8b and 8c showing the effect of parallelism on search latency as the number of documents increases. For this experiment, you need all instances except `baseline-client` and `baseline-server` running (so `server-1`-`server-8`, `master`, and `client`). Run the following commands locally:

```
cd bench
python3 exp_fig8b-c.py      # 18 minutes
python3 plot_fig8b.py       # few seconds
python3 plot_fig8c.py       # few seconds
```

This will produce plots close to Figures 8b and 8c on page 11 of the paper in `bench/out/fig8b.png` and `bench/out/fig8c.png`. Note that these plotting scripts use the data we collected for the baseline (in `bench/ref`) rather than experimental data, and we show how to validate the data we collected for the baseline at a reduced scale later.

<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig8b.png" width="400">
Figure 8b

<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig8c.png" width="400">
Figure 8c
### Figures 10-11

Run the experiment and then plot the data for Figures 10 and 11 showing the effect of parallelism on throughput as the number of documents increases for different workloads. For this experiment, you need all instances except `baseline-client` and `baseline-server` running (so `server-1`-`server-8`, `master`, and `client`). Run the following commands locally:

```
cd bench
python3 exp_fig10-11.py     # 40 minutes
python3 plot_fig10a.py      # few seconds
python3 plot_fig10b.py      # few seconds
python3 plot_fig10c.py      # few seconds
python3 plot_fig11a.py      # few seconds
python3 plot_fig11b.py      # few seconds
python3 plot_fig11c.py      # few seconds
```

This will produce plots close to Figures 10 and 11 on page 11 of the paper in `bench/out/fig10a.png`, `bench/out/fig10b.png`, `bench/out/fig10c.png`, `bench/out/fig11a.png`, `bench/out/fig11b.png`, `bench/out/fig11c.png`. Again, these plotting scripts use the data collected for the baseline (in `bench/ref`) rather than experimental data, and we show how to validate the data we collected for the baseline at a reduced scale next.

<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig10a.png" width="400">
Figure 10a
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig10b.png" width="400">
Figure 10b
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig10c.png" width="400">
Figure 10c
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig11a.png" width="400">
Figure 11a
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig11b.png" width="400">
Figure 11b
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig11c.png" width="400">
Figure 11c

### Baseline

To validate the baseline results we used for the above figures, we show how to reproduce our baseline results for 1,024 and 2,048 documents. This process takes several hours (whereas collecting all the data points takes approximately a week).

For this experiment, you only need `baseline-client` and baseline-server` running. We do not run this experiment locally (broken SSH pipes over a long period of time make this more challenging to script). Instead, you must SSH into `baseline-client` directly. From `baseline-client`, you will need to SSH into `baseline-server`, which you can do by copying your SSH key to `baseline-client` or via SSH agent forwarding.

Run the following commands on `baseline-client` (recommend doing this as a background task or in a `tmux` session because it takes a long time):
```
cd dory/baseline
./runTests.sh <server-IP-addr> <path-to-SSH-key>   # TODO: Add time estimate
```

TODO: explain how to compare to output

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
