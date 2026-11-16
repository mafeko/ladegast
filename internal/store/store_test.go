package store

import (
	"os"
	"path/filepath"
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

func TestGuestCRUD(t *testing.T) {
	s := openTestStore(t)

	f := Guest{ID: "guest-1", Name: "Alex", VehicleName: "vehicle_1", PINHash: "hash"}
	if err := s.CreateGuest(f); err != nil {
		t.Fatalf("CreateGuest: %v", err)
	}
	if err := s.CreateGuest(f); err == nil {
		t.Fatal("expected error creating duplicate guest, got nil")
	}

	got, err := s.GetGuest("guest-1")
	if err != nil {
		t.Fatalf("GetGuest: %v", err)
	}
	if got.Name != "Alex" {
		t.Errorf("Name = %q, want Alex", got.Name)
	}

	if _, err := s.GetGuest("missing"); err != ErrNotFound {
		t.Errorf("GetGuest(missing) err = %v, want ErrNotFound", err)
	}

	f.Name = "Alexandra"
	if err := s.UpdateGuest(f); err != nil {
		t.Fatalf("UpdateGuest: %v", err)
	}
	got, _ = s.GetGuest("guest-1")
	if got.Name != "Alexandra" {
		t.Errorf("Name after update = %q, want Alexandra", got.Name)
	}

	if err := s.UpdateGuest(Guest{ID: "missing"}); err != ErrNotFound {
		t.Errorf("UpdateGuest(missing) err = %v, want ErrNotFound", err)
	}

	list, err := s.ListGuests()
	if err != nil {
		t.Fatalf("ListGuests: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("len(list) = %d, want 1", len(list))
	}
}

func TestListGuests_CorruptData(t *testing.T) {
	s := openTestStore(t)
	err := s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucketGuests)).Put([]byte("broken"), []byte("not json"))
	})
	if err != nil {
		t.Fatalf("seed corrupt data: %v", err)
	}
	if _, err := s.ListGuests(); err == nil {
		t.Fatal("expected error listing guests with corrupt data, got nil")
	}
}

func TestDeleteGuest(t *testing.T) {
	s := openTestStore(t)
	f := Guest{ID: "guest-1", Name: "Alex"}
	if err := s.CreateGuest(f); err != nil {
		t.Fatalf("CreateGuest: %v", err)
	}
	if err := s.DeleteGuest("guest-1"); err != nil {
		t.Fatalf("DeleteGuest: %v", err)
	}
	if _, err := s.GetGuest("guest-1"); err != ErrNotFound {
		t.Errorf("GetGuest after delete err = %v, want ErrNotFound", err)
	}
	if err := s.DeleteGuest("guest-1"); err != ErrNotFound {
		t.Errorf("DeleteGuest(missing) err = %v, want ErrNotFound", err)
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

func TestOpen_FailsWhenDBPathIsADirectory(t *testing.T) {
	dir := t.TempDir()
	// Pre-create "ladegast.db" as a directory so bolt.Open fails trying to
	// open it as a file.
	if err := os.Mkdir(dir+"/ladegast.db", 0o700); err != nil {
		t.Fatalf("os.Mkdir: %v", err)
	}
	if _, err := Open(dir); err == nil {
		t.Fatal("expected error opening bbolt file that is actually a directory, got nil")
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
	key := "1.2.3.4:guest-1"
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
	key := "1.2.3.4:guest-1"

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

func TestOpen_MigratesLegacyFriendsBucket(t *testing.T) {
	dir := t.TempDir()
	db, err := bolt.Open(filepath.Join(dir, "ladegast.db"), 0o600, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("friends"))
		if err != nil {
			return err
		}
		return b.Put([]byte("id1"), []byte("legacy-value"))
	})
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	err = s.db.View(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte("friends")) != nil {
			t.Error("legacy bucket should be gone")
		}
		if got := tx.Bucket([]byte(bucketGuests)).Get([]byte("id1")); string(got) != "legacy-value" {
			t.Errorf("migrated value = %q", got)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
