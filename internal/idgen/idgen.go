// Package idgen creates the random identifiers Ladegast hands to friends:
// UUIDs for permanent login links and numeric PINs.
package idgen

import (
	"crypto/rand"
	"fmt"
)

// UUID returns a random RFC 4122 version 4 UUID.
func UUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Sprintf("idgen: read random bytes: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// PIN returns a random numeric PIN of the given length (e.g. 6 -> 6 digits,
// zero-padded, "000000".."999999").
func PIN(length int) string {
	max := int64(1)
	for i := 0; i < length; i++ {
		max *= 10
	}
	n, err := randInt64(max)
	if err != nil {
		panic(fmt.Sprintf("idgen: generate PIN: %v", err))
	}
	return fmt.Sprintf("%0*d", length, n)
}

func randInt64(max int64) (int64, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, err
	}
	var n uint64
	for _, v := range b {
		n = n<<8 | uint64(v)
	}
	return int64(n % uint64(max)), nil
}
