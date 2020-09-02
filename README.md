# DORY: An Encrypted Search System with Distributed Trust

DORY is an encrypted search system that splits trust between multiple servers in order to efficiently hide access patterns from a malicious attacker who controls all but one of the servers. This implementation contains our DORY search protocol as described in the OSDI20 paper and does not include a complementary end-to-end encrypted filesystem that could use or interface with DORY.

This implementation accompanies our paper "DORY: An Encrypted Search System with Distributed Trust" by Emma Dauterman, Eric Feng, Ellen Luo, Raluca Ada Popa, and Ion Stoica to appear at OSDI20

**WARNING**: This is an academic proof-of-concept prototype and has not received careful code review. This implementation is NOT ready for production use.

This prototype is released under the Apache v2 license (see [License](#license)).

## Setup

For our experiments, we will use a cluster of AWS EC2 instances. Reviewers should have been provided with credentials to our AWS environment with compute resources.

1. [2 minutes] Make sure python3 is downloaded. Then run the following:
```
git clone https://github.com/ucbrise/dory
cd dory/bench
mkdir out
pip3 install -r requirements.txt

```

2. [5 minutes] Install [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) (version 2 works) and run `aws configure` using the instructions [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html) (use `json` as the default output format, and it does not matter what default region you choose).

3. [2 minutes] To start a cluster, run the following:
```
cd bench
python3 start_cluster.py
```
This will create the EC2 instances for the experiments using the correct AMI and copy configuration files to each instance. Default TLS keys and certificates are included for testing. You do not need to change these to run evaluation benchmarks, but in a real deployment, these should be freshly generated for security.

Note: If you see a message that a SSH connection was refused on port 22, then the script was not able to copy over the configuration file because the instance had not fully started yet. In this case, either teardown the cluster and restart (waiting a few minutes between teardown and starting again), or manually copy the configuration files yourself using `scp`.

4. When you are done with experiments (or just done for the day), run the following to terminate all instances:
```
cd bench
python3 teardown_cluster.py
```
Make sure to wait a few minutes between tearing down one cluster and setting up another cluster. It takes a few minutes for instances to fully terminate, and if you try to set up another cluster before the other has been fully shut down, you will exceed your vCPU limits.

You've just finished setup! Follow the steps below to run experiments and reproduce our results.


## Running experiments

The experimental results in this paper compare DORY to a PathORAM baseline in `baseline/`. Unfortunately, running the experiments to produce the data in our paper takes about a week. We will show how to validate our baseline results for a small number of documents and then for the other figures, we will use the results we produced for the baseline in order to reproduce the figures in our paper.

With the exception of the baseline experiment, all experiments produce `.dat` files with raw data in `bench/out` and figures corresponding to the ones in the paper as `.png` files in `bench/out`. The corresponding raw data and figures we produced for the paper are in `bench/ref` for comparison.

The experiments for Table 7, Figures 8b-8c, and Figures 10-11 cannot be run concurrently. However, the experiments for the baseline can be run at the same time as the DORY experiments (we recommend doing this to save time, as the baseline experiments take a few hours to complete).

To speed up testing, some of the experiments start with an index that is built by the server where the server has the keys to generate a correct search index. This configuration should only be used for testing (for security, only the client should have the keys).

### Baseline

To validate the baseline results we used for the above figures, we show how to reproduce our baseline results for 1,024 and 2,048 documents. This process takes several hours (whereas collecting all the data points takes approximately a week).

Run the following commands to start the experiment:
```
cd bench 
python3 start_baseline.py   # 1 minute
```

This script starts the experiment and returns immediately. In approximately 70 minutes, retrieve the results by running: 
```
cd bench
python3 get_baseline_results.py     # Run 70 minutes after start_baseline.py
```
This will copy the output of the baseline experiments to `dory/baseline/out/oram_1024` and `dory/baseline/out/oram_2048`.

Compare the reported search latency to the search latency points in Figure 8b, or look at the exact data points in `bench/ref/latency_search_oram.dat` (all data points reported in milliseconds). Compare the throughput for different workloads to the throughput points in Figures 10a, 10b, and 10c, or look at the exact data points in `bench/ref/oram_throughput_1_9.dat`, `bench/ref/oram_throughput_5_5.dat`, and `bench/ref/oram_throughput_9_1.dat` (units are operations per second).

### Table 7

Run the experiment to collect the data for the part of Table 7 showing the breakdown of search latency. For this experiment, you only need to have `server-1`, `server-2`, `master`, and `client` running. Run the following commands locally:

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

Figure 8b:
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig8b.png" width="400">

Figure 8c:
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig8c.png" width="400">

### Figures 10-11

Run the experiment and then plot the data for Figures 10 and 11 showing the effect of parallelism on throughput as the number of documents increases for different workloads. For this experiment, you need all instances except `baseline-client` and `baseline-server` running (so `server-1`-`server-8`, `master`, and `client`). Run the following commands locally:

```
cd bench
python3 exp_fig10-11.py     # 123 minutes
python3 plot_fig10a.py      # few seconds
python3 plot_fig10b.py      # few seconds
python3 plot_fig10c.py      # few seconds
python3 plot_fig11a.py      # few seconds
python3 plot_fig11b.py      # few seconds
python3 plot_fig11c.py      # few seconds
```

This will produce plots close to Figures 10 and 11 on page 11 of the paper in `bench/out/fig10a.png`, `bench/out/fig10b.png`, `bench/out/fig10c.png`, `bench/out/fig11a.png`, `bench/out/fig11b.png`, `bench/out/fig11c.png`. Again, these plotting scripts use the data collected for the baseline (in `bench/ref`) rather than experimental data, and we show how to validate the data we collected for the baseline at a reduced scale next. You may see some variation in comparison to the graphs from the paper because we do not average over multiple trials in order to save time.

Figure 10a:
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig10a.png" width="400">

Figure 10b:
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig10b.png" width="400">

Figure 10c:
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig10c.png" width="400">

Figure 11a:
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig11a.png" width="400">

Figure 11b:
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig11b.png" width="400">

Figure 11c:
<img src="https://github.com/ucbrise/dory/blob/master/bench/ref/fig11c.png" width="400">


## Test and play with functionality

Outside of replicating our results, you can also run correctness tests and interactively search for keywords over a set of sample documents. You can do this remotely or locally.

### Building from source

If installing from source, follow the below instructions:
1. Install OpenSSL, tested up to version 2.6.5.
2. Run `go get github.com/hashicorp/go-msgpack/codec`.
3. Download and build `libstemmer` (http://snowball.tartarus.org/download.html), tested up to version 2.0.0.
4. Move the output `libstemmer.o` to `src/c/`.
5. In `bench/` run `pip3 install -r requirements.txt`.

### Local configuration

To configure DORY to run locally, run the following:

```
cd bench
python3 start_local.py
```

You can then start the master, servers, and client on your local machine.

### Running DORY 
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

### Tests
To test the low-level crypto, run `make` in `src/c` and run `correctness_tests`. To test the end-to-end system, run the client with the correctness test flag set to true (`-c`).

## Acknowledgements

We build on Saba Eskandarian's DPF implementation in https://github.com/SabaEskandarian/Express.
