// Package idgen creates the random identifiers Ladegast hands to friends:
// UUIDs for permanent login links and numeric PINs.
package idgen

import (
	"crypto/rand"
	"fmt"

	"github.com/google/uuid"
)

// UUID returns a random RFC 4122 version 4 UUID. Generation and validation
// are delegated to google/uuid - the de facto standard library for this in
// Go, since the standard library itself has no UUID package.
func UUID() string {
	return uuid.NewString()
}

// ValidUUID reports whether s is a syntactically valid UUID, regardless of
// version or variant.
func ValidUUID(s string) bool {
	return uuid.Validate(s) == nil
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
