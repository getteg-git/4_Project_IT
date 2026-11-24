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

func TestValidateUserPayload_AllowsTechnicianWithoutDepartment(t *testing.T) {
	err := validateUserPayload("tech01", "password", "Test Technician", "tech@example.com", "technician", nil, []int{1}, false)
	if err != nil {
		t.Fatalf("expected technician without a department to be allowed, got: %v", err)
	}
}

func TestValidateUserPayload_StillRequiresTechnicianSpecialtyWithoutDepartment(t *testing.T) {
	err := validateUserPayload("tech02", "password", "Test Technician", "tech2@example.com", "technician", nil, nil, false)
	if err == nil {
		t.Fatal("expected technicians to need at least one specialty")
	}
}
