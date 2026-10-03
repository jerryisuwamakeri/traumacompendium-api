package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/certificate"
)

// beginnerLevelID is Level 1's fixed ID, seeded by migration 0014 — used
// here only to satisfy certificates' FK on levels, not to exercise
// learning content itself.
const beginnerLevelID = "00000000-0000-0000-0000-000000000001"

func TestCertificateRepository_CreateThenFind(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewCertificateRepository(db)
	ctx := context.Background()

	issuedAt := time.Now().UTC().Truncate(time.Second)
	c := &certificate.Certificate{
		ID:       uuid.NewString(),
		UserID:   userID,
		LevelID:  beginnerLevelID,
		Serial:   "TC-TESTSERIAL01",
		IssuedAt: issuedAt,
	}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	byUserLevel, err := repo.FindByUserAndLevel(ctx, userID, beginnerLevelID)
	if err != nil {
		t.Fatalf("FindByUserAndLevel() error = %v", err)
	}
	if byUserLevel.Serial != "TC-TESTSERIAL01" {
		t.Errorf("Serial = %q, want %q", byUserLevel.Serial, "TC-TESTSERIAL01")
	}
	if !byUserLevel.IssuedAt.Equal(issuedAt) {
		t.Errorf("IssuedAt = %v, want %v", byUserLevel.IssuedAt, issuedAt)
	}

	bySerial, err := repo.FindBySerial(ctx, "TC-TESTSERIAL01")
	if err != nil {
		t.Fatalf("FindBySerial() error = %v", err)
	}
	if bySerial.UserID != userID {
		t.Errorf("UserID = %q, want %q", bySerial.UserID, userID)
	}
}

func TestCertificateRepository_FindByUserAndLevel_NotFound(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewCertificateRepository(db)

	_, err := repo.FindByUserAndLevel(context.Background(), userID, beginnerLevelID)
	if !errors.Is(err, certificate.ErrNotFound) {
		t.Errorf("FindByUserAndLevel() for a never-issued certificate error = %v, want %v", err, certificate.ErrNotFound)
	}
}

func TestCertificateRepository_FindBySerial_NotFound(t *testing.T) {
	db := openTestDB(t)
	repo := NewCertificateRepository(db)

	_, err := repo.FindBySerial(context.Background(), "TC-DOESNOTEXIST")
	if !errors.Is(err, certificate.ErrNotFound) {
		t.Errorf("FindBySerial() for an unknown serial error = %v, want %v", err, certificate.ErrNotFound)
	}
}

func TestCertificateRepository_Create_DuplicateUserLevel_Fails(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewCertificateRepository(db)
	ctx := context.Background()

	first := &certificate.Certificate{
		ID: uuid.NewString(), UserID: userID, LevelID: beginnerLevelID,
		Serial: "TC-FIRST0001", IssuedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("Create(first) error = %v", err)
	}

	second := &certificate.Certificate{
		ID: uuid.NewString(), UserID: userID, LevelID: beginnerLevelID,
		Serial: "TC-SECOND0001", IssuedAt: time.Now().UTC(),
	}
	if err := repo.Create(ctx, second); err == nil {
		t.Error("Create(second) for the same user+level succeeded, want a unique-key error")
	}
}
