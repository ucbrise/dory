source CONFIG.mine

max_docs="1024"
bf_sz="1120"
correct="false"

while getopts ":h?:d:n:t:b:s:c:" opt; do
    case "$opt" in
        h|\?)
            echo -e "\nArguments: "
            echo -e "-b \t\t Bits in Bloom filter (default 1120)"
            echo -e "-n \t\t Max number of documents (default 1024)"
            echo -e "-s \t\t Server number (default 1)"
            echo -e "-m \t\t Malicious security? (default true)"
            exit 0
            ;;
        t)
            correct=$OPTARG
            ;;
        n)      
            max_docs=$OPTARG
            ;;  
        b)      
            bf_sz=$OPTARG
            ;;  
        s)
            server_num=$OPTARG
            ;;
        c)
            correct=$OPTARG
            ;;
    esac        
done    

echo "bench_dir='$bench_dir', bf_sz='$bf_sz', max_docs='$max_docs', server_num='$server_num'"

CGO_CFLAGS="-I"$OSSL"/include" CGO_LDFLAGS="-L"$OSSL"/lib -lcrypto -lm "$PWD"/src/c/libstemmer.o" go run src/server/server.go --config=src/config/server"$server_num".config --bf_sz="$bf_sz" --max_docs="$max_docs"
