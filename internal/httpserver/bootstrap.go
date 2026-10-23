package httpserver

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/mafeko/ladegast/internal/config"
	"github.com/mafeko/ladegast/internal/store"
)

// EnsureAdminBootstrap creates the initial admin account from
// LADEGAST_ADMIN_USERNAME/LADEGAST_ADMIN_PASSWORD if no admin exists yet.
// It is a no-op once at least one admin account has been created, even if
// those variables are still set (they should be removed from the
// environment after the first successful boot).
func EnsureAdminBootstrap(st *store.Store, cfg config.Config) error {
	hasAdmin, err := st.HasAnyAdmin()
	if err != nil {
		return fmt.Errorf("check existing admins: %w", err)
	}
	if hasAdmin {
		return nil
	}
	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		return fmt.Errorf("no admin account exists and LADEGAST_ADMIN_USERNAME/LADEGAST_ADMIN_PASSWORD are not set")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	return st.CreateAdmin(store.Admin{Username: cfg.AdminUsername, PasswordHash: string(hash)})
}
