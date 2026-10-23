package store

import (
	"encoding/json"

	bolt "go.etcd.io/bbolt"
)

// Admin is the single operator account managing friends. Multiple admin
// accounts are supported by the schema even though the MVP provisions
// exactly one at first boot.
type Admin struct {
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
}

func (s *Store) CreateAdmin(a Admin) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketAdmins))
		return putJSON(b, a.Username, a)
	})
}

func (s *Store) GetAdmin(username string) (Admin, error) {
	var a Admin
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketAdmins))
		v := b.Get([]byte(username))
		if v == nil {
			return ErrNotFound
		}
		return json.Unmarshal(v, &a)
	})
	return a, err
}

// HasAnyAdmin reports whether at least one admin account exists.
func (s *Store) HasAnyAdmin() (bool, error) {
	found := false
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketAdmins))
		k, _ := b.Cursor().First()
		found = k != nil
		return nil
	})
	return found, err
}
