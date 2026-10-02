package safereachservice

import (
	"context"
	"testing"

	"github.com/rs/zerolog"

	"github.com/traumacompendium/backend/internal/domain/safereach"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
	"github.com/traumacompendium/backend/internal/testsupport/fakes"
)

func newTestService() *Service {
	return New(
		fakes.NewSafeReachContactRepo(),
		fakes.NewSafeReachAlertRepo(),
		fakes.NewDeviceTokenRepo(),
		&fakes.SMSSender{},
		&fakes.EmailSender{},
		&fakes.PushSender{},
		&fakes.Publisher{},
		"https://app.traumacompendium.test",
		zerolog.Nop(),
	)
}

func TestRegisterContact_Success(t *testing.T) {
	svc := newTestService()

	c, err := svc.RegisterContact(context.Background(), "user-1", RegisterContactInput{
		Name: "Ada Okafor", Relationship: "Sister", Phone: "+2348012345678",
	})
	if err != nil {
		t.Fatalf("RegisterContact() error = %v", err)
	}
	if c.UserID != "user-1" || c.Name != "Ada Okafor" {
		t.Errorf("RegisterContact() = %+v, want UserID=user-1 Name=Ada Okafor", c)
	}
}

func TestRegisterContact_StripsHTMLFromNameAndRelationship(t *testing.T) {
	svc := newTestService()

	c, err := svc.RegisterContact(context.Background(), "user-1", RegisterContactInput{
		Name: `<script>alert(1)</script>Ada`, Relationship: "<b>Sister</b>", Phone: "+2348012345678",
	})
	if err != nil {
		t.Fatalf("RegisterContact() error = %v", err)
	}
	if c.Name != "Ada" {
		t.Errorf("Name = %q, want HTML stripped", c.Name)
	}
	if c.Relationship != "Sister" {
		t.Errorf("Relationship = %q, want HTML stripped", c.Relationship)
	}
}

func TestRegisterContact_MissingName_ReturnsValidationError(t *testing.T) {
	svc := newTestService()

	_, err := svc.RegisterContact(context.Background(), "user-1", RegisterContactInput{Phone: "+2348012345678"})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeValidation {
		t.Errorf("RegisterContact() with no name error = %v, want validation_error", err)
	}
}

func TestRegisterContact_NoPhoneOrEmail_ReturnsValidationError(t *testing.T) {
	svc := newTestService()

	_, err := svc.RegisterContact(context.Background(), "user-1", RegisterContactInput{Name: "Ada"})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeValidation {
		t.Errorf("RegisterContact() with no phone/email error = %v, want validation_error", err)
	}
}

func TestRegisterContact_AtMaxContacts_ReturnsUnprocessable(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	for i := 0; i < safereach.MaxTrustedContacts; i++ {
		_, err := svc.RegisterContact(ctx, "user-1", RegisterContactInput{Name: "Contact", Phone: "+2348000000000"})
		if err != nil {
			t.Fatalf("RegisterContact() #%d error = %v", i, err)
		}
	}

	_, err := svc.RegisterContact(ctx, "user-1", RegisterContactInput{Name: "One Too Many", Phone: "+2348011111111"})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeUnprocessable {
		t.Errorf("RegisterContact() beyond the limit error = %v, want unprocessable_entity", err)
	}
}

func TestListContacts_OnlyReturnsCallersContacts(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	if _, err := svc.RegisterContact(ctx, "user-1", RegisterContactInput{Name: "A", Phone: "+2348011111111"}); err != nil {
		t.Fatalf("RegisterContact() error = %v", err)
	}
	if _, err := svc.RegisterContact(ctx, "user-2", RegisterContactInput{Name: "B", Phone: "+2348022222222"}); err != nil {
		t.Fatalf("RegisterContact() error = %v", err)
	}

	list, err := svc.ListContacts(ctx, "user-1")
	if err != nil {
		t.Fatalf("ListContacts() error = %v", err)
	}
	if len(list) != 1 || list[0].Name != "A" {
		t.Errorf("ListContacts(user-1) = %+v, want just contact A", list)
	}
}

func TestDeleteContact_WrongUser_ReturnsNotFound(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	c, err := svc.RegisterContact(ctx, "owner", RegisterContactInput{Name: "A", Phone: "+2348011111111"})
	if err != nil {
		t.Fatalf("RegisterContact() error = %v", err)
	}

	err = svc.DeleteContact(ctx, "intruder", c.ID)
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Errorf("DeleteContact() by non-owner error = %v, want not_found", err)
	}
}

func TestDeleteContact_Success_RemovesFromList(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	c, err := svc.RegisterContact(ctx, "user-1", RegisterContactInput{Name: "A", Phone: "+2348011111111"})
	if err != nil {
		t.Fatalf("RegisterContact() error = %v", err)
	}

	if err := svc.DeleteContact(ctx, "user-1", c.ID); err != nil {
		t.Fatalf("DeleteContact() error = %v", err)
	}

	list, err := svc.ListContacts(ctx, "user-1")
	if err != nil {
		t.Fatalf("ListContacts() error = %v", err)
	}
	if len(list) != 0 {
		t.Errorf("ListContacts() after delete = %+v, want empty", list)
	}
}
