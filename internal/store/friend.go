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
	ID                string       `json:"id"`
	Name              string       `json:"name"`
	VehicleName       string       `json:"vehicleName"` // EVCC config name
	PINHash           string       `json:"pinHash"`
	RateSolarPerKWh   float64      `json:"rateSolarPerKWh"`
	RateGridPerKWh    float64      `json:"rateGridPerKWh"`
	CreatedAt         time.Time    `json:"createdAt"`
	LinkRegeneratedAt time.Time    `json:"linkRegeneratedAt"`
	Settlements       []Settlement `json:"settlements,omitempty"`
}

// Settlement marks the moment a friend paid for all charging up to At. Only
// the timestamp is stored; amounts are derived from EVCC sessions between
// consecutive settlements. Stored inside the Friend so a regenerated link
// carries its history along and deleting the friend removes it.
type Settlement struct {
	ID string    `json:"id"`
	At time.Time `json:"at"`
}

// AddSettlement appends a settlement to the friend's record.
func (s *Store) AddSettlement(friendID string, st Settlement) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketFriends))
		v := b.Get([]byte(friendID))
		if v == nil {
			return ErrNotFound
		}
		var f Friend
		if err := json.Unmarshal(v, &f); err != nil {
			return err
		}
		f.Settlements = append(f.Settlements, st)
		return putJSON(b, f.ID, f)
	})
}

// DeleteSettlement removes one settlement; ErrNotFound if friend or
// settlement does not exist.
func (s *Store) DeleteSettlement(friendID, settlementID string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketFriends))
		v := b.Get([]byte(friendID))
		if v == nil {
			return ErrNotFound
		}
		var f Friend
		if err := json.Unmarshal(v, &f); err != nil {
			return err
		}
		kept := f.Settlements[:0]
		for _, st := range f.Settlements {
			if st.ID != settlementID {
				kept = append(kept, st)
			}
		}
		if len(kept) == len(f.Settlements) {
			return ErrNotFound
		}
		f.Settlements = kept
		return putJSON(b, f.ID, f)
	})
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
