// Package store persists Ladegast's own entities (friends, admins, login
// rate-limiting) in a local bbolt file. Charging data itself is never
// stored here — it is read live from EVCC (see internal/evcc).
package store

import (
	"fmt"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
)

var buckets = []string{bucketFriends, bucketAdmins, bucketRateLimit}

const (
	bucketFriends   = "friends"
	bucketAdmins    = "admins"
	bucketRateLimit = "rate_limit"
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
