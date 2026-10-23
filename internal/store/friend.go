package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

var ErrNotFound = errors.New("not found")

// Friend is a person allowed to charge their vehicle, identified by a
// permanent link (ID, used as the URL slug) protected by a PIN.
type Friend struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	VehicleName       string    `json:"vehicleName"` // EVCC config name
	PINHash           string    `json:"pinHash"`
	RateSolarPerKWh   float64   `json:"rateSolarPerKWh"`
	RateGridPerKWh    float64   `json:"rateGridPerKWh"`
	CreatedAt         time.Time `json:"createdAt"`
	LinkRegeneratedAt time.Time `json:"linkRegeneratedAt"`
}

func (s *Store) CreateFriend(f Friend) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketFriends))
		if b.Get([]byte(f.ID)) != nil {
			return fmt.Errorf("friend %q already exists", f.ID)
		}
		return putJSON(b, f.ID, f)
	})
}

func (s *Store) UpdateFriend(f Friend) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketFriends))
		if b.Get([]byte(f.ID)) == nil {
			return ErrNotFound
		}
		return putJSON(b, f.ID, f)
	})
}

func (s *Store) GetFriend(id string) (Friend, error) {
	var f Friend
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketFriends))
		v := b.Get([]byte(id))
		if v == nil {
			return ErrNotFound
		}
		return json.Unmarshal(v, &f)
	})
	return f, err
}

func (s *Store) DeleteFriend(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketFriends))
		if b.Get([]byte(id)) == nil {
			return ErrNotFound
		}
		return b.Delete([]byte(id))
	})
}

func (s *Store) ListFriends() ([]Friend, error) {
	var out []Friend
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketFriends))
		return b.ForEach(func(_, v []byte) error {
			var f Friend
			if err := json.Unmarshal(v, &f); err != nil {
				return err
			}
			out = append(out, f)
			return nil
		})
	})
	return out, err
}

func putJSON(b *bolt.Bucket, key string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return b.Put([]byte(key), data)
}
