package user

import "testing"

func TestValidateUserPayload_AllowsEmptyPasswordForEdit(t *testing.T) {
	err := validateUserPayload("admin01", "", "Test User", "test@example.com", "admin", nil, nil, true)
	if err != nil {
		t.Fatalf("expected empty password to be allowed during updates, got: %v", err)
	}
}

func TestValidateUserPayload_RejectsPasswordOutsideRange(t *testing.T) {
	err := validateUserPayload("admin01", "a", "Test User", "test@example.com", "admin", nil, nil, false)
	if err == nil {
		t.Fatal("expected password length 3-30 to be enforced")
	}
}
