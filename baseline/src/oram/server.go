/*
 * Implementation of the server in Path ORAM by Stefanov et. al.
 */
package oram

import (
	"math"
	// "net"
)

/*
 * Each bucket contains Z blocks; it suffices that Z is small e.g. 4
 */
type Bucket []Block

/*
 * The server
 */
type Server struct {
	N     int    // total number of blocks outsourced
	L     int    // height of binary tree
	B     int    // block size in bytes, currently fixed at 32
	Z     int    // capacity of each bucket in blocks
    tree  [][][]byte
}

/*
 * Initialize a Server
 *
 * Returns a new Server with params:
 *   N: Number of blocks outsourced to the server
 *   B: Capacity of each block in bytes (fixed to 32 bytes)
 *   Z: Capacity of each bucket in blocks
 */
func InitServer(N int, Z int, B int) *Server {
    B = B + 28      // add tag and nonce
    s := &Server{N: N, B: B, Z: Z}
	// height of tree: log2(N)
	s.L = int(math.Ceil(math.Log2(float64(N))))

    s.tree = make([][][]byte, s.L + 1)
    for i := 0; i <= s.L; i++ {
        s.tree[i] = make([][]byte, s.N/2)
        for j := 0; j < s.N/2; j++ {
            s.tree[i][j] = make([]byte, s.B * s.Z)
        }
    }

	return s
}


func (s *Server) WriteNode(b Bucket, l int, n int) {
	max_n := (1 << uint(l))
	if n < 0 || n >= max_n {
		return
	}

	// get raw bytes of bucket
	bucket_bytes := bucket_join(b, nil)

	//level, off := s.foffset(l, n)
    level := l
    off := n / s.Z

    copy(s.tree[level][off], bucket_bytes)
}

func (s *Server) ReadNode(l int, n int) (Bucket, error) {
	// get file and offset into that file
	//level, offset := s.foffset(l, n)
    level := l
    offset := n / s.Z

	// in bytes
	bucket_size := s.B * s.Z

	// read all bytes at once
	buf := make([]byte, bucket_size)
    copy(buf, s.tree[level][offset])

	// organize bytes into buckets
	bucket := make(Bucket, s.Z)
	for i := range bucket {
		left := i * s.B
		right := (i + 1) * s.B
		bucket[i] = buf[left:right]
	}

	return bucket, nil
}

// returns the file and an offset into that file for a bucket at a given level
func (s *Server) foffset(l int, n int) (int, int) {
	// bounds check
	if n < 0 || n >= (1<<uint(l)) {
		return l, 0
	}

	total_bytes := s.B * s.Z * n
	off := total_bytes % (s.Z * s.B * s.N / 2) 

	return l, off
}
