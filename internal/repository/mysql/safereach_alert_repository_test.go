package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/safereach"
	"github.com/traumacompendium/backend/internal/pkg/hash"
)

func newTestAlert(t *testing.T, userID string) *safereach.Alert {
	t.Helper()
	return &safereach.Alert{
		ID:             uuid.NewString(),
		UserID:         userID,
		Level:          safereach.LevelUrgent,
		Status:         safereach.StatusActive,
		Message:        "Need help getting home safely.",
		ShareTokenHash: hash.Token(uuid.NewString()),
		TriggeredAt:    time.Now().UTC().Truncate(time.Second),
	}
}

func TestSafeReachAlertRepository_CreateFindUpdate(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewSafeReachAlertRepository(db)
	ctx := context.Background()

	a := newTestAlert(t, userID)
	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Status != safereach.StatusActive || got.Level != safereach.LevelUrgent {
		t.Errorf("FindByID() = %+v, want Status=active Level=urgent", got)
	}

	ownedGot, err := repo.FindOwnedByID(ctx, a.ID, userID)
	if err != nil {
		t.Fatalf("FindOwnedByID() error = %v", err)
	}
	if ownedGot.ID != a.ID {
		t.Errorf("FindOwnedByID() ID = %q, want %q", ownedGot.ID, a.ID)
	}

	resolvedAt := time.Now().UTC().Truncate(time.Second)
	a.Status = safereach.StatusResolved
	a.ResolvedAt = &resolvedAt
	if err := repo.Update(ctx, a); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, _ = repo.FindByID(ctx, a.ID)
	if got.Status != safereach.StatusResolved || got.ResolvedAt == nil {
		t.Errorf("after Update() = %+v, want Status=resolved with ResolvedAt set", got)
	}
}

func TestSafeReachAlertRepository_FindOwnedByID_WrongUser_NotFound(t *testing.T) {
	db := openTestDB(t)
	owner := seedTestUser(t, db)
	intruder := seedTestUser(t, db)
	repo := NewSafeReachAlertRepository(db)
	ctx := context.Background()

	a := newTestAlert(t, owner)
	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err := repo.FindOwnedByID(ctx, a.ID, intruder)
	if !errors.Is(err, safereach.ErrAlertNotFound) {
		t.Errorf("FindOwnedByID() by non-owner error = %v, want %v", err, safereach.ErrAlertNotFound)
	}

	// FindByID is deliberately unscoped (used by the public share-link
	// view), so the non-owner CAN read it by ID directly.
	got, err := repo.FindByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("FindByID() (unscoped) error = %v", err)
	}
	if got.ID != a.ID {
		t.Errorf("FindByID() ID = %q, want %q", got.ID, a.ID)
	}
}

func TestSafeReachAlertRepository_Update_WrongUser_NotFound(t *testing.T) {
	db := openTestDB(t)
	owner := seedTestUser(t, db)
	intruder := seedTestUser(t, db)
	repo := NewSafeReachAlertRepository(db)
	ctx := context.Background()

	a := newTestAlert(t, owner)
	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	impostor := *a
	impostor.UserID = intruder
	impostor.Status = safereach.StatusCancelled

	err := repo.Update(ctx, &impostor)
	if !errors.Is(err, safereach.ErrAlertNotFound) {
		t.Errorf("Update() by non-owner error = %v, want %v", err, safereach.ErrAlertNotFound)
	}
}

func TestSafeReachAlertRepository_List_PaginatesNewestFirst(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewSafeReachAlertRepository(db)
	ctx := context.Background()

	base := time.Now().UTC().Add(-3 * time.Hour).Truncate(time.Second)
	for i := 0; i < 3; i++ {
		a := newTestAlert(t, userID)
		a.TriggeredAt = base.Add(time.Duration(i) * time.Hour)
		if err := repo.Create(ctx, a); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	alerts, total, err := repo.List(ctx, userID, safereach.ListFilter{Limit: 2})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if len(alerts) != 2 {
		t.Fatalf("len(alerts) = %d, want 2", len(alerts))
	}
	if !alerts[0].TriggeredAt.After(alerts[1].TriggeredAt) {
		t.Error("List() did not return alerts newest-first")
	}
}

func TestSafeReachAlertRepository_AddLocationUpdate_AndLatestLocation(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewSafeReachAlertRepository(db)
	ctx := context.Background()

	a := newTestAlert(t, userID)
	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	none, err := repo.LatestLocation(ctx, a.ID)
	if err != nil {
		t.Fatalf("LatestLocation() before any update error = %v", err)
	}
	if none != nil {
		t.Errorf("LatestLocation() before any update = %+v, want nil", none)
	}

	base := time.Now().UTC().Truncate(time.Second)
	for i, coords := range [][2]float64{{6.5244, 3.3792}, {6.5300, 3.3800}} {
		u := &safereach.LocationUpdate{
			ID: uuid.NewString(), AlertID: a.ID,
			Lat: coords[0], Lng: coords[1],
			RecordedAt: base.Add(time.Duration(i) * time.Minute),
		}
		if err := repo.AddLocationUpdate(ctx, u); err != nil {
			t.Fatalf("AddLocationUpdate() error = %v", err)
		}
	}

	latest, err := repo.LatestLocation(ctx, a.ID)
	if err != nil {
		t.Fatalf("LatestLocation() error = %v", err)
	}
	if latest == nil {
		t.Fatal("LatestLocation() = nil, want the most recent fix")
	}
	if latest.Lat != 6.5300 || latest.Lng != 3.3800 {
		t.Errorf("LatestLocation() = (%v, %v), want (6.53, 3.38)", latest.Lat, latest.Lng)
	}
}
