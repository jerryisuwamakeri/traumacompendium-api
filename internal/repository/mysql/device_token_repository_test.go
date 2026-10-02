package mysql

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/traumacompendium/backend/internal/domain/devicetoken"
)

func TestDeviceTokenRepository_RegisterAndList(t *testing.T) {
	db := openTestDB(t)
	userID := seedTestUser(t, db)
	repo := NewDeviceTokenRepository(db)
	ctx := context.Background()

	d := &devicetoken.DeviceToken{ID: uuid.NewString(), UserID: userID, Token: "fcm-token-abc", Platform: "android"}
	if err := repo.Register(ctx, d); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	list, err := repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("ListByUser() error = %v", err)
	}
	if len(list) != 1 || list[0].Token != "fcm-token-abc" {
		t.Errorf("ListByUser() = %+v, want one token fcm-token-abc", list)
	}
}

func TestDeviceTokenRepository_Register_ReRegisteringSameToken_Reassigns(t *testing.T) {
	db := openTestDB(t)
	userA := seedTestUser(t, db)
	userB := seedTestUser(t, db)
	repo := NewDeviceTokenRepository(db)
	ctx := context.Background()

	tokenID := uuid.NewString()
	if err := repo.Register(ctx, &devicetoken.DeviceToken{ID: tokenID, UserID: userA, Token: "shared-device", Platform: "ios"}); err != nil {
		t.Fatalf("Register() first error = %v", err)
	}
	// Same physical device, different account signs in: re-registering
	// the same token must repoint it, not error on the unique constraint.
	if err := repo.Register(ctx, &devicetoken.DeviceToken{ID: uuid.NewString(), UserID: userB, Token: "shared-device", Platform: "ios"}); err != nil {
		t.Fatalf("Register() re-registration error = %v", err)
	}

	listA, err := repo.ListByUser(ctx, userA)
	if err != nil {
		t.Fatalf("ListByUser(userA) error = %v", err)
	}
	if len(listA) != 0 {
		t.Errorf("ListByUser(userA) = %+v, want empty after reassignment", listA)
	}

	listB, err := repo.ListByUser(ctx, userB)
	if err != nil {
		t.Fatalf("ListByUser(userB) error = %v", err)
	}
	if len(listB) != 1 || listB[0].Token != "shared-device" {
		t.Errorf("ListByUser(userB) = %+v, want the reassigned token", listB)
	}
}
