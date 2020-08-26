/*
 * Test functions for the oram2pc library
 */
package oram

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func Benchmark_encryption(b *testing.B) {
	// set up 
	server := "test"
	N := 4096
	L := int(math.Ceil(math.Log2(float64(N))))
	Z := 4
	fsize := 4096
	c := InitClient(N, Z)
	c.AddServer(server, N, Z, fsize)

	key, _ := c.keys[server]

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for i := 0; i < L; i++ {
			r := rand.Int31n(int32(N))
			val := uint64(r)
			id := int(r)
			blk := block_encode(id, val)
			enc := enc_block(blk, key)
			_ = dec_block(enc, key)
		}
	}
}

func Benchmark_readpath(b *testing.B) {
	// set up 
	server := "test"
	N := 4096
	Z := 4
	fsize := 4096
	c := InitClient(N, Z)
	c.AddServer(server, N, Z, fsize)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		r := int(rand.Int31n(int32(N)))
		_, err := c.get_path_buckets(r)
		if err != nil {
			panic(err)
		}
	}

	// old_path, err := s.get_path(x)
}

func Benchmark_readwritepath(b *testing.B) {
	// set up 
	server := "test"
	N := 4096
	Z := 4
	fsize := 4096
	c := InitClient(N, Z)
	c.AddServer(server, N, Z, fsize)

	s := c.server
	key, _ := c.keys[server]

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r := int(rand.Int31n(int32(N)))
		path, err := c.get_path(r)
		if err != nil {
			panic(err)
		}

		for i := range path {
			cur_l := len(path) - 1 - i
			bucket := make_bucket(nil, Z, key)
			s.WriteNode(bucket, cur_l, path[i])
		}
	}
}

func Benchmark_randomwrite(b *testing.B) {
	// set up 
	s := "test"
	N := 4096
	Z := 4
	fsize := 4096
	c := InitClient(N, Z)
	c.AddServer(s, N, Z, fsize)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// write to a random block
		r := rand.Int31n(int32(N))
		_, err := c.Access(s, true, int(r), uint64(r))
		if err != nil {
			panic(err)
		}
	}

	c.RemoveServer(s)
}

func Benchmark_sequentialwrite(b *testing.B) {
	// set up 
	s := "test"
	N := 4096
	Z := 4
	fsize := 4096
	c := InitClient(N, Z)
	c.AddServer(s, N, Z, fsize)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// write to every block
		_, err := c.Access(s, true, n % N, uint64(n % N))
		if err != nil {
			panic(err)
		}
	}

	c.RemoveServer(s)
}

func Test_blocks(t *testing.T) {
	key := []byte("key lul")

	aoeu := enc_dummy_block(key)
	aoeu2 := enc_dummy_block(key)
	aoeu3 := enc_dummy_block(key)
	aoeu4 := enc_dummy_block(key)
	b := enc_block(block_encode(0x1234, 0x1122334455667788), key)
	fmt.Println(block_encode(0x1234, 0x1122334455667788))
	fmt.Println(b)
	id, d, dummy := block_decode(dec_block(b, key))
	if dummy == false {
		fmt.Printf("%x: %x\n", id, d)
	}

	_, e, dummy := block_decode(dec_block(aoeu, key))
	if dummy == true {
		fmt.Println("found dummy block!", e)
	}

	bucket := Bucket{b, aoeu2, aoeu3, aoeu4}
	dummy_bucket := Bucket{aoeu, aoeu2, aoeu3, aoeu4}
	joined := bucket_join(bucket, nil)
	fmt.Println(aoeu)
	fmt.Println(joined[:32])
	fmt.Println(aoeu2)
	fmt.Println(joined[32:64])
	fmt.Println(aoeu3)
	fmt.Println(joined[64:96])
	fmt.Println(aoeu4)
	fmt.Println(joined[96:])
	if len(joined) != len(bucket[0])*4 {
		panic("Check bucket_join!!!")
	}

	bux := []Bucket{dummy_bucket, dummy_bucket, dummy_bucket, bucket}
	fmt.Println(find_nondummy(bux, key))

	idx, val := bucket_find_block(bux, 0x1234, key)
	fmt.Println("Finding nondummy in buckets: index", idx, "val", val)

	bucket2 := make_bucket([]Block{aoeu}, 4, key)
	fmt.Println(bucket2)
}

func Test_client(t *testing.T) {
	// currently use 256-bit blocks to store encrypted 64-bit values
	c := InitClient(4, 4)

	c.AddServer("test", c.N, c.Z, 4096)
	fmt.Println(c.ServerInfo("test"))

	buckets, err := c.get_path_buckets(2)
	if err != nil {
		panic(err)
	}

	for i := range buckets {
		fmt.Println("Reading bucket", strconv.Itoa(i))
		for j := range buckets[i] {
			fmt.Println("Reading block", strconv.Itoa(j))
			fmt.Println(buckets[i][j])
		}
	}

	fmt.Println("Trying to write to a = 0")
	_, err = c.Access("test", true, 0, uint64(0xdeadbeef))
	if err != nil {
		panic(err)
	}
	fmt.Println("Trying to read from a = 0")
	val, err := c.Access("test", false, 0, uint64(0xdeadbeef))
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
	fmt.Println("Trying to read from a = 0")
	val, err = c.Access("test", false, 0, uint64(0xdeadbeef))
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
	fmt.Println("Trying to read from a = 0")
	val, err = c.Access("test", false, 0, uint64(0xdeadbeef))
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
	fmt.Println("Trying to read from a = 0")
	val, err = c.Access("test", false, 0, uint64(0xdeadbeef))
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
	fmt.Println("Trying to write to a = 1")
	_, err = c.Access("test", true, 1, uint64(0x10))
	if err != nil {
		panic(err)
	}
	fmt.Println("Trying to read from a = 0")
	val, err = c.Access("test", false, 0, uint64(0xdeadbeef))
	if err != nil {
		panic(err)
	}
	fmt.Println(val)

	c.RemoveServer("test")
}
