package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/traumacompendium/backend/internal/pkg/apperror"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestFail_KnownAppError_MapsStatusAndCode(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Fail(c, apperror.NewConflict("email already registered"))

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", w.Code, http.StatusConflict)
	}

	var body envelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.Error == nil {
		t.Fatal("expected an error body")
	}
	if body.Error.Code != apperror.CodeConflict {
		t.Errorf("Code = %q, want %q", body.Error.Code, apperror.CodeConflict)
	}
	if body.Data != nil {
		t.Error("expected no data field alongside an error")
	}
}

func TestFail_UnknownError_Returns500WithoutLeakingDetail(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	Fail(c, errors.New("connection refused to internal-db-host:5432: password authentication failed"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var body envelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.Error.Code != apperror.CodeInternal {
		t.Errorf("Code = %q, want %q", body.Error.Code, apperror.CodeInternal)
	}
	if body.Error.Message != "an unexpected error occurred" {
		t.Errorf("Message = %q, leaked internal detail", body.Error.Message)
	}
}

func TestBindJSON_ValidationFailure_ReturnsFieldErrors(t *testing.T) {
	type input struct {
		Email string `json:"email" binding:"required,email"`
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"not-an-email"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	var dst input
	if BindJSON(c, &dst) {
		t.Fatal("BindJSON() = true, want false for an invalid email")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var body envelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if body.Error.Fields["Email"] == "" {
		t.Error("expected a field-level message for Email")
	}
}

func TestBindJSON_Valid_ReturnsTrue(t *testing.T) {
	type input struct {
		Email string `json:"email" binding:"required,email"`
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"jane@example.com"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	var dst input
	if !BindJSON(c, &dst) {
		t.Fatal("BindJSON() = false, want true for a valid payload")
	}
	if dst.Email != "jane@example.com" {
		t.Errorf("Email = %q, want %q", dst.Email, "jane@example.com")
	}
}
