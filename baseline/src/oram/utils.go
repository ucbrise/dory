/*
 * Useful functions that aren't specific to a client or server
 */

package oram

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
    "crypto/aes"
    "crypto/cipher"
    "io"
    "encoding/base64"
	"encoding/binary"
    "fmt"
	"log"
	"math"
	"math/big"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func xor_bytes(a []byte, b []byte) []byte {
	if len(a) != len(b) {
		log.Fatal("XOR_bytes: arguments are not the same length!")
	}

	r := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		r[i] = a[i] ^ b[i]
	}

	return r
}

func random_perm(size int64) []int64 {
	p := make([]int64, size, size)

	var i int64
	for i = 0; i < size; i++ {
		p[i] = i
	}

	// use fisher-yates shuffling
	for i = 0; i < size-2; i++ {
		j_big, err := rand.Int(rand.Reader, big.NewInt(size-1-i))
		if err != nil {
			log.Println(err)
		}
		j := j_big.Int64()

		p[i], p[j] = p[j], p[i]
	}

	return p
}

// generate a uint32 in [0, max)
func gen_uint32(max uint32) uint32 {
	for {
		// get random bytes
		max_f := float64(max - 1)
		num_bits := uint(math.Ceil(math.Log2(max_f)))

		r := make([]byte, 4)
		_, err := rand.Read(r)
		if err != nil {
			log.Println(err)
		}

		// trim bytes to get the right number of bits
		var extra_bits uint
		extra_bits = 4*8 - num_bits
		guess := binary.LittleEndian.Uint32(r) >> extra_bits

		if guess < max {
			return guess
		}
	}
}

// this is a stupid hack but I'm lazy
func gen_int(max int) int {
	return int(gen_uint32(uint32(max)))
}

func gen_alphanum_string(size uint8) string {
	b := make([]byte, size)
	for i := range b {
		b[i] = letters[gen_uint32(uint32(len(letters)))]
	}

	return string(b)
}

func pad(v []byte, c byte) []byte {
	// right pad to 128 bits
	p := make([]byte, 16)
	copy(p, v)
	for i := len(v); i < len(p); i++ {
		p[i] = c
	}

	return p
}

func unpad(v []byte, c byte) []byte {
	// undo right pad
	padding := make([]byte, 1)
	padding[0] = c
	return bytes.Replace(v, padding, nil, len(v))
}

// padding and b64 encoding for plaintext, pad to 128 bits
func pt_encode(v []byte) []byte {
	encoded_len := base64.RawStdEncoding.EncodedLen(len(v))
    b64_v := make([]byte, encoded_len)
	base64.RawStdEncoding.Encode(b64_v, v)

	return pad(b64_v, 0x24)
}

func pt_decode(v []byte) []byte {
	unpadded := unpad(v, 0x24)
	decoded_len := base64.RawStdEncoding.DecodedLen(len(unpadded))
	decoded_v := make([]byte, decoded_len)
	_, err := base64.RawStdEncoding.Decode(decoded_v, unpadded)
	if err != nil {
		return nil
	}

	return decoded_v
}

// dunno if this should be done this way but whatever
func prf(k []byte, r []byte) []byte {
	return sha256.New().Sum(append(k, r...))
}

// multi-message secure encryption defined in Pass & Shelat 3.7 (pg 94)
func encrypt(m []byte, k []byte) []byte {
/*	// generate random string r of desired length
	r := make([]byte, len(m))
	n, err := rand.Read(r)
	if n != len(r) || err != nil {
		return nil
	}

	// xor with PRF(r)
	xor_part := xor_bytes(m, prf(k, r)[:len(r)])

    fmt.Println("len of msg: ", len(m))
	cip := append(r, xor_part...)
	return cip
*/
    block,_ := aes.NewCipher(k)
    nonce := make([]byte, 12)
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil
    }

    aesgcm,_ := cipher.NewGCM(block)
    ct := aesgcm.Seal(nil, nonce, m, nil)
    fullCt := append(nonce, ct...)
    return fullCt
}

func decrypt(cip []byte, k []byte) []byte {
    block,_ := aes.NewCipher(k)
    aesgcm,_ := cipher.NewGCM(block)
    pt, err := aesgcm.Open(nil, cip[:12], cip[12:], nil)
    if err != nil {
        fmt.Println("DECRYPTION FAILED")
        return nil
    }
    return pt

    /*r := cip[
    :len(cip)/2]
	xor_part := cip[len(cip)/2:]

	m := xor_bytes(xor_part, prf(k, r)[:len(r)])

	return m*/
}
