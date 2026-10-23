package store

import (
	"os"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestFriendCRUD(t *testing.T) {
	s := openTestStore(t)

	f := Friend{ID: "friend-1", Name: "Alex", VehicleName: "vehicle_1", PINHash: "hash"}
	if err := s.CreateFriend(f); err != nil {
		t.Fatalf("CreateFriend: %v", err)
	}
	if err := s.CreateFriend(f); err == nil {
		t.Fatal("expected error creating duplicate friend, got nil")
	}

	got, err := s.GetFriend("friend-1")
	if err != nil {
		t.Fatalf("GetFriend: %v", err)
	}
	if got.Name != "Alex" {
		t.Errorf("Name = %q, want Alex", got.Name)
	}

	if _, err := s.GetFriend("missing"); err != ErrNotFound {
		t.Errorf("GetFriend(missing) err = %v, want ErrNotFound", err)
	}

	f.Name = "Alexandra"
	if err := s.UpdateFriend(f); err != nil {
		t.Fatalf("UpdateFriend: %v", err)
	}
	got, _ = s.GetFriend("friend-1")
	if got.Name != "Alexandra" {
		t.Errorf("Name after update = %q, want Alexandra", got.Name)
	}

	if err := s.UpdateFriend(Friend{ID: "missing"}); err != ErrNotFound {
		t.Errorf("UpdateFriend(missing) err = %v, want ErrNotFound", err)
	}

	list, err := s.ListFriends()
	if err != nil {
		t.Fatalf("ListFriends: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("len(list) = %d, want 1", len(list))
	}
}

func TestDeleteFriend(t *testing.T) {
	s := openTestStore(t)
	f := Friend{ID: "friend-1", Name: "Alex"}
	if err := s.CreateFriend(f); err != nil {
		t.Fatalf("CreateFriend: %v", err)
	}
	if err := s.DeleteFriend("friend-1"); err != nil {
		t.Fatalf("DeleteFriend: %v", err)
	}
	if _, err := s.GetFriend("friend-1"); err != ErrNotFound {
		t.Errorf("GetFriend after delete err = %v, want ErrNotFound", err)
	}
	if err := s.DeleteFriend("friend-1"); err != ErrNotFound {
		t.Errorf("DeleteFriend(missing) err = %v, want ErrNotFound", err)
	}
}

func TestOpen_FailsWhenDataDirIsAFile(t *testing.T) {
	dir := t.TempDir()
	blocked := dir + "/blocked"
	if err := os.WriteFile(blocked, []byte("x"), 0o600); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	// blocked exists as a regular file, so MkdirAll(blocked) must fail.
	if _, err := Open(blocked); err == nil {
		t.Fatal("expected error opening store in a path blocked by a file, got nil")
	}
}

func TestAdminCRUD(t *testing.T) {
	s := openTestStore(t)

	ok, err := s.HasAnyAdmin()
	if err != nil {
		t.Fatalf("HasAnyAdmin: %v", err)
	}
	if ok {
		t.Fatal("HasAnyAdmin = true before any admin created")
	}

	if err := s.CreateAdmin(Admin{Username: "admin", PasswordHash: "hash"}); err != nil {
		t.Fatalf("CreateAdmin: %v", err)
	}

	ok, err = s.HasAnyAdmin()
	if err != nil {
		t.Fatalf("HasAnyAdmin: %v", err)
	}
	if !ok {
		t.Fatal("HasAnyAdmin = false after creating one")
	}

	a, err := s.GetAdmin("admin")
	if err != nil {
		t.Fatalf("GetAdmin: %v", err)
	}
	if a.PasswordHash != "hash" {
		t.Errorf("PasswordHash = %q, want hash", a.PasswordHash)
	}

	if _, err := s.GetAdmin("missing"); err != ErrNotFound {
		t.Errorf("GetAdmin(missing) err = %v, want ErrNotFound", err)
	}
}

func TestRateLimit(t *testing.T) {
	s := openTestStore(t)
	key := "1.2.3.4:friend-1"
	now := time.Now()
	window := 15 * time.Minute

	count, err := s.CountRecentFailures(key, now, window)
	if err != nil {
		t.Fatalf("CountRecentFailures: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}

	for i := 0; i < 3; i++ {
		if err := s.RecordFailedAttempt(key, now, window); err != nil {
			t.Fatalf("RecordFailedAttempt: %v", err)
		}
	}
	count, err = s.CountRecentFailures(key, now, window)
	if err != nil {
		t.Fatalf("CountRecentFailures: %v", err)
	}
	if count != 3 {
		t.Fatalf("count = %d, want 3", count)
	}

	// Attempts outside the window are pruned.
	count, err = s.CountRecentFailures(key, now.Add(window+time.Second), window)
	if err != nil {
		t.Fatalf("CountRecentFailures: %v", err)
	}
	if count != 0 {
		t.Fatalf("count after window elapsed = %d, want 0", count)
	}

	if err := s.RecordFailedAttempt(key, now, window); err != nil {
		t.Fatalf("RecordFailedAttempt: %v", err)
	}
	if err := s.ResetAttempts(key); err != nil {
		t.Fatalf("ResetAttempts: %v", err)
	}
	count, _ = s.CountRecentFailures(key, now, window)
	if count != 0 {
		t.Fatalf("count after reset = %d, want 0", count)
	}
}

func TestCountRecentFailures_CorruptData(t *testing.T) {
	s := openTestStore(t)
	key := "1.2.3.4:friend-1"

	err := s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketRateLimit)).Put([]byte(key), []byte("not json"))
	})
	if err != nil {
		t.Fatalf("seed corrupt data: %v", err)
	}

	if _, err := s.CountRecentFailures(key, time.Now(), 15*time.Minute); err == nil {
		t.Fatal("expected error reading corrupt rate-limit data, got nil")
	}
}
