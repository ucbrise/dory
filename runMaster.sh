max_docs="128"
bf_sz="1024"
malicious="true"
tick_ms="1000"
num_clusters="1"

while getopts ":h?:d:n:t:b:m:c:" opt; do
    case "$opt" in
        h|\?)
            echo -e "\n Arguments: "
            echo -e "-b \t\t Bits in Bloom filter (default 1120)"
            echo -e "-n \t\t Max number of documents (default 1024)"
            echo -e "-m \t\t Malicious security? (default true)"
            echo -e "-t \t\t Milliseconds between master -> server updates (default 1000)"
            echo -e "-p \t\t Number of clusters (default 1)"
            exit 0
            ;;
        n)      
            max_docs=$OPTARG
            ;;  
        b)      
            bf_sz=$OPTARG
            ;;
        m)
            malicious=$OPTARG
            ;;
        t)
            tick_ms=$OPTARG
            ;;
        p)
            num_clusters=$OPTARG
            ;;
    esac        
done    

echo "bench_dir='$bench_dir', bf_sz='$bf_sz', max_docs='$max_docs', malicious='$malicious'"


CGO_CFLAGS="-I./libsolv-sys/src -D LIBSOLV_INTERNAL -w" CGO_LDFLAGS="-lssl -lcrypto -lm -lpthread "$PWD"/src/c/libstemmer.o" go run src/master/master.go --config=src/config/master.config --bf_sz="$bf_sz" --max_docs="$max_docs" --malicious="$malicious" --tick_ms="$tick_ms" --num_clusters="$num_clusters"
