// Package store persists Ladegast's own entities (guests, admins, login
// rate-limiting) in a local bbolt file. Charging data itself is never
// stored here — it is read live from EVCC (see internal/evcc).
package store

import (
	"fmt"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
)

var buckets = []string{bucketGuests, bucketAdmins, bucketRateLimit}

const (
	bucketGuests    = "guests"
	bucketAdmins    = "admins"
	bucketRateLimit = "rate_limit"

	// legacyBucketGuests is the pre-rename name of bucketGuests; existing
	// databases are migrated on Open.
	legacyBucketGuests = "friends"
)

type Store struct {
	db *bolt.DB
}

// Open opens (creating if necessary) the bbolt database file inside dataDir.
func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	db, err := bolt.Open(filepath.Join(dataDir, "ladegast.db"), 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		if err := migrateLegacyGuests(tx); err != nil {
			return err
		}
		for _, name := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(name)); err != nil {
				return fmt.Errorf("create bucket %q: %w", name, err)
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// migrateLegacyGuests moves the contents of the old "friends" bucket into
// "guests" (once) so databases created before the rename keep their data.
func migrateLegacyGuests(tx *bolt.Tx) error {
	old := tx.Bucket([]byte(legacyBucketGuests))
	if old == nil {
		return nil
	}
	if tx.Bucket([]byte(bucketGuests)) != nil {
		return nil // already migrated; leave the legacy bucket untouched
	}
	dst, err := tx.CreateBucket([]byte(bucketGuests))
	if err != nil {
		return fmt.Errorf("create bucket %q: %w", bucketGuests, err)
	}
	if err := old.ForEach(func(k, v []byte) error { return dst.Put(k, v) }); err != nil {
		return fmt.Errorf("migrate guests: %w", err)
	}
	return tx.DeleteBucket([]byte(legacyBucketGuests))
}
