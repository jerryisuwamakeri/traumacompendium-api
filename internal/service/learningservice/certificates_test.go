package learningservice

import (
	"context"
	"testing"

	"github.com/traumacompendium/backend/internal/domain/user"
	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

func TestCompleteLesson_LastOfLevel_IssuesCertificate(t *testing.T) {
	svc, content, _, _, _, certificates, users := newTestServiceWithCertificates()
	levelID, lesson1ID, lesson2ID, _, _, _, _ := seedTwoLessonLevel(content)
	users.Seed(&user.User{ID: "user-1", FullName: "Ada Lovelace", Email: "ada@example.com", Phone: "+1000000001"})
	ctx := context.Background()

	if _, err := svc.CompleteLesson(ctx, "user-1", lesson1ID, ""); err != nil {
		t.Fatalf("CompleteLesson(lesson1) error = %v", err)
	}
	result, err := svc.CompleteLesson(ctx, "user-1", lesson2ID, "")
	if err != nil {
		t.Fatalf("CompleteLesson(lesson2) error = %v", err)
	}
	if !result.CertificateIssued {
		t.Error("CertificateIssued = false, want true on first full completion")
	}

	cert, err := certificates.FindByUserAndLevel(ctx, "user-1", levelID)
	if err != nil {
		t.Fatalf("FindByUserAndLevel() error = %v", err)
	}
	if cert.Serial == "" {
		t.Error("issued certificate has an empty serial")
	}
}

func TestCompleteLesson_ReCompleting_DoesNotReissueCertificate(t *testing.T) {
	svc, content, _, _, _, certificates, users := newTestServiceWithCertificates()
	_, lesson1ID, lesson2ID, _, _, _, _ := seedTwoLessonLevel(content)
	users.Seed(&user.User{ID: "user-1", FullName: "Ada Lovelace", Email: "ada@example.com", Phone: "+1000000001"})
	ctx := context.Background()

	if _, err := svc.CompleteLesson(ctx, "user-1", lesson1ID, ""); err != nil {
		t.Fatalf("CompleteLesson(lesson1) error = %v", err)
	}
	if _, err := svc.CompleteLesson(ctx, "user-1", lesson2ID, ""); err != nil {
		t.Fatalf("CompleteLesson(lesson2) error = %v", err)
	}
	firstCert, err := certificates.FindByUserAndLevel(ctx, "user-1", "level-1")
	if err != nil {
		t.Fatalf("FindByUserAndLevel() error = %v", err)
	}

	result, err := svc.CompleteLesson(ctx, "user-1", lesson2ID, "")
	if err != nil {
		t.Fatalf("CompleteLesson(lesson2, again) error = %v", err)
	}
	if result.CertificateIssued {
		t.Error("CertificateIssued = true on a re-completion, want false (already issued)")
	}

	secondCert, err := certificates.FindByUserAndLevel(ctx, "user-1", "level-1")
	if err != nil {
		t.Fatalf("FindByUserAndLevel() error = %v", err)
	}
	if firstCert.Serial != secondCert.Serial {
		t.Errorf("serial changed across re-completion: %q -> %q", firstCert.Serial, secondCert.Serial)
	}
}

func TestGetCertificate_NotYetEarned_NotFound(t *testing.T) {
	svc, content, _, _, _, _, users := newTestServiceWithCertificates()
	levelID, _, _, _, _, _, _ := seedTwoLessonLevel(content)
	users.Seed(&user.User{ID: "user-1", FullName: "Ada Lovelace", Email: "ada@example.com", Phone: "+1000000001"})

	_, err := svc.GetCertificate(context.Background(), "user-1", levelID, "")
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("err = %v, want a NotFound apperror", err)
	}
}

func TestGetCertificate_Earned_RendersPDFWithRecipientName(t *testing.T) {
	svc, content, _, _, _, _, users := newTestServiceWithCertificates()
	levelID, lesson1ID, lesson2ID, _, _, _, _ := seedTwoLessonLevel(content)
	users.Seed(&user.User{ID: "user-1", FullName: "Ada Lovelace", Email: "ada@example.com", Phone: "+1000000001"})
	ctx := context.Background()

	if _, err := svc.CompleteLesson(ctx, "user-1", lesson1ID, ""); err != nil {
		t.Fatalf("CompleteLesson(lesson1) error = %v", err)
	}
	if _, err := svc.CompleteLesson(ctx, "user-1", lesson2ID, ""); err != nil {
		t.Fatalf("CompleteLesson(lesson2) error = %v", err)
	}

	pdfBytes, err := svc.GetCertificate(ctx, "user-1", levelID, "")
	if err != nil {
		t.Fatalf("GetCertificate() error = %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatal("GetCertificate() returned an empty PDF")
	}
	if got := string(pdfBytes[:5]); got != "%PDF-" {
		t.Errorf("PDF header = %q, want %%PDF-", got)
	}
}

func TestVerifyCertificate_UnknownSerial_ReturnsInvalid(t *testing.T) {
	svc, _, _, _, _, _, _ := newTestServiceWithCertificates()

	result, err := svc.VerifyCertificate(context.Background(), "TC-DOESNOTEXIST")
	if err != nil {
		t.Fatalf("VerifyCertificate() error = %v", err)
	}
	if result.Valid {
		t.Error("Valid = true for an unknown serial, want false")
	}
}

func TestVerifyCertificate_KnownSerial_ReturnsDetails(t *testing.T) {
	svc, content, _, _, _, _, users := newTestServiceWithCertificates()
	levelID, lesson1ID, lesson2ID, _, _, _, _ := seedTwoLessonLevel(content)
	users.Seed(&user.User{ID: "user-1", FullName: "Ada Lovelace", Email: "ada@example.com", Phone: "+1000000001"})
	ctx := context.Background()

	if _, err := svc.CompleteLesson(ctx, "user-1", lesson1ID, ""); err != nil {
		t.Fatalf("CompleteLesson(lesson1) error = %v", err)
	}
	if _, err := svc.CompleteLesson(ctx, "user-1", lesson2ID, ""); err != nil {
		t.Fatalf("CompleteLesson(lesson2) error = %v", err)
	}

	cert, err := svc.certificates.FindByUserAndLevel(ctx, "user-1", levelID)
	if err != nil {
		t.Fatalf("FindByUserAndLevel() error = %v", err)
	}

	result, err := svc.VerifyCertificate(ctx, cert.Serial)
	if err != nil {
		t.Fatalf("VerifyCertificate() error = %v", err)
	}
	if !result.Valid || result.UserName != "Ada Lovelace" || result.LevelName != "Beginner" {
		t.Errorf("result = %+v, want valid=true, user=Ada Lovelace, level=Beginner", result)
	}
}
