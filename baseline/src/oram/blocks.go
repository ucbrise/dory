/*
 * Implements various functions related to blocks
 */

package oram

import (
	"bytes"
)

type Block []byte

// makes a bucket consisting of blks and padded with dummy blocks until
// the bucket reaches a size of max
func make_bucket(blks []Block, max int, key []byte, B int) Bucket {

	bucket := make(Bucket, max)

	end := len(blks)
	if end > max {
		end = max;
	}

	for i := 0; i < end; i++ {
		bucket[i] = enc_block(blks[i], key)
	}

	// pad with encrypted dummy blocks
	for i := end; i < max; i++ {
		bucket[i] = enc_dummy_block(key, B)
	}

	return bucket
}

// splits a bucket into its plaintext blocks (opposite of make_bucket)
func split_bucket(bucket Bucket, key []byte) []Block {
	blocks := make([]Block, len(bucket))
	for i := range bucket {
		blocks[i] = dec_block(bucket[i], key)
	}

	return blocks
}

// finds all non-dummy blocks in some buckets
func find_nondummy(bux []Bucket, key []byte) []Block {
	nondummy := make([]Block, len(bux)*len(bux[0]))
	num_nd := 0
	for i := range bux {
		for j := range bux[i] {
			cur_blk := dec_block(bux[i][j], key)
			if !is_dummy(cur_blk) {
				nondummy[num_nd] = cur_blk
				num_nd += 1
				continue
			}
		}
	}

	return nondummy[:num_nd]
}

// find which bucket the block with id "id" is found
func bucket_find_block(bux []Bucket, id int, key []byte) (int, []byte) {
	for i := range bux {
		for j := range bux[i] {
			cur_blk := bux[i][j]

			if !is_dummy(dec_block(cur_blk, key)) {
				return i, dec_block(cur_blk, key)
			}
		}
	}

	return -1, []byte{}
}

// do the same thing as bucket_find_block but in a slice of blocks
func slice_find_block(blocks []Block, id int) int {
	for i := range blocks {
		if !is_dummy(blocks[i]){
			return i
		}
	}

	return -1
}

// Join concatenates the elements of s to create a new byte slice. The separator
// sep is placed between elements in the resulting slice.
func bucket_join(s Bucket, sep []byte) []byte {

	if len(s) == 0 {
		return []byte{}
	}

	if len(s) == 1 {
		// Just return a copy.
		return append([]byte(nil), s[0]...)
	}

	n := len(sep) * (len(s) - 1)
	for _, v := range s {
		n += len(v)
	}

	b := make([]byte, n)
	bp := copy(b, s[0])
	for _, v := range s[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], v)
	}

	return b
}

/*
 * Returns an unencrypted dummy block
 */
func dummy_block(B int) Block {
	dummy := make([]byte, B)
	for i := range dummy {
		dummy[i] = 0xff
	}

	return dummy
}

/*
 * Detects whether the byte slice is the unencrypted dummy block
 */
func is_dummy(blk Block) bool {
	result := bytes.Compare(blk, dummy_block(len(blk)))
	return result == 0
}

/*
 * Returns an encrypted version of the dummy block which has the format:
 * | 0xFFFF... |
 * <- 128 bits ->
 */
func enc_dummy_block(k []byte, B int) Block {
	dummy_plain := dummy_block(B)
	dummy_cip := encrypt(dummy_plain, k)

	return dummy_cip
}

/*
 * Returns an encrypted version of the encoded block
 */
func enc_block(blk Block, k []byte) Block {
	return Block(encrypt([]byte(blk), k))
}

/*
 * Returns the plaintext (a uint64) by decrypting an encrypted block if
 * the bool == false, else the block is a dummy block
 */
func dec_block(blk Block, k []byte) Block {
	return Block(decrypt([]byte(blk), k))
}
