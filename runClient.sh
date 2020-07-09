source CONFIG.mine

bench_dir=""
correct="false"
bf_sz="1120"
num_docs="1024"
malicious="true"
fast_setup="true"
use_master="true"
throughput="false"
throughput_sec="60"
throughput_threads="64"
num_updates="5"
num_searches="5"
num_clusters="0"
only_setup="false"

while getopts ":h?:d:t:b:n:m:f:s:c:x:y:q:r:p:z:" opt; do
    case "$opt" in
        h|\?)
            echo -e "\nArguments: "
            echo -e "-b \t\t Bits in Bloom filter (default 1120)"
            echo -e "-n \t\t Max number of documents (default 1024)"
            echo -e "-m \t\t Malicious security? (default true)"
            echo -e "-s \t\t Run with master? (default true)"
            echo -e "-p \t\t Number of clusters (default 1)"
            echo -e "\nBenchmarking/testing arguments: "
            echo -e "-c \t\t Run correctness tests? (default false)"
            echo -e "-f \t\t Run with fast setup? (default true; insecure, only for benchmarking)"
            echo -e "-t \t\t Run throughput tests? (default false)"
            echo -e "-x \t\t Seconds to run throughput tests (default 60)"
            echo -e "-y \t\t Client threads for throughput tests (default 64)"
            echo -e "-q \t\t Number of consecutive updates in throughput tests (default 5)"
            echo -e "-r \t\t Number of consecutive searches in throughput tests (default 5)"
            echo -e "-d \t\t Input directory for generating updates for benchmarks\n"
            exit 0
            ;;
        d)
            bench_dir=$OPTARG
            ;;
        c)
            correct=$OPTARG
            ;;
        b)
            bf_sz=$OPTARG
            ;;
        n)
            num_docs=$OPTARG
            ;;
        m)
            malicious=$OPTARG
            ;;

        f)  fast_setup=$OPTARG
            ;;

        s)  use_master=$OPTARG
            ;;
        
        t) throughput=$OPTARG
            ;;
        
        x) throughput_sec=$OPTARG
            ;;

        y) throughput_threads=$OPTARG
            ;;
        
        q) num_updates=$OPTARG
            ;;

        r) num_searches=$OPTARG
            ;;

        p) num_clusters=$OPTARG
            ;;

        z) only_setup=$OPTARG
            ;;

    esac
done

echo "bench_dir='$bench_dir', tests='$correct', bf_sz='$bf_sz', num_docs='$num_docs'"

CGO_LDFLAGS="-lssl -lpthread -lcrypto -lm "$PWD"/src/c/libstemmer.o" go run src/bench/client.go --config=src/config/client.config --test="$correct" --bench_dir="$bench_dir" --bf_sz="$bf_sz" --num_docs="$num_docs" --malicious="$malicious" --fast_setup="$fast_setup" --use_master="$use_master" --throughput="$throughput" --throughput_sec="$throughput_sec" --throughput_threads="$throughput_threads" --num_updates="$num_updates" --num_searches="$num_searches" --num_clusters="$num_clusters" --only_setup="$only_setup"

