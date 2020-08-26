package oram

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func Test_utils(t *testing.T) {
	m := make([]byte, 4)
	rand.Read(m)
	fmt.Println("Random test subject:", m)

	fmt.Println("Testing pad functions:")
	fmt.Println(pad(m, 0x24))
	fmt.Println(unpad(pad(m, 0x24), 0x24))

	fmt.Println("Testing plaintext encoding and decoding functions:")
	fmt.Println(pt_encode(m))
	fmt.Println(pt_decode(pt_encode(m)))

	k := make([]byte, 128/8)
	rand.Read(k)
	cip := encrypt(m, k)
	fmt.Println("Testing encrypting and decrypting...")
	fmt.Println(cip)
	fmt.Println(decrypt(cip, k))
}
