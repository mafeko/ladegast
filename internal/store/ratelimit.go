package store

import (
	"encoding/json"
	"time"

	bolt "go.etcd.io/bbolt"
)

// RecordFailedAttempt appends a failed-login timestamp under key (typically
// "<client-ip>:<friend-id>") and prunes entries older than window in the
// same transaction.
func (s *Store) RecordFailedAttempt(key string, now time.Time, window time.Duration) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketRateLimit))
		attempts, err := readAttempts(b, key)
		if err != nil {
			return err
		}
		attempts = pruneAttempts(attempts, now, window)
		attempts = append(attempts, now)
		return putJSON(b, key, attempts)
	})
}

// CountRecentFailures returns how many failed attempts under key happened
// within window before now, pruning older entries as a side effect.
func (s *Store) CountRecentFailures(key string, now time.Time, window time.Duration) (int, error) {
	count := 0
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketRateLimit))
		attempts, err := readAttempts(b, key)
		if err != nil {
			return err
		}
		pruned := pruneAttempts(attempts, now, window)
		count = len(pruned)
		if len(pruned) != len(attempts) {
			return putJSON(b, key, pruned)
		}
		return nil
	})
	return count, err
}

// ResetAttempts clears recorded failures for key, e.g. after a successful login.
func (s *Store) ResetAttempts(key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketRateLimit))
		return b.Delete([]byte(key))
	})
}

func readAttempts(b *bolt.Bucket, key string) ([]time.Time, error) {
	v := b.Get([]byte(key))
	if v == nil {
		return nil, nil
	}
	var attempts []time.Time
	if err := json.Unmarshal(v, &attempts); err != nil {
		return nil, err
	}
	return attempts, nil
}

func pruneAttempts(attempts []time.Time, now time.Time, window time.Duration) []time.Time {
	cutoff := now.Add(-window)
	out := attempts[:0:0]
	for _, t := range attempts {
		if t.After(cutoff) {
			out = append(out, t)
		}
	}
	return out
}
