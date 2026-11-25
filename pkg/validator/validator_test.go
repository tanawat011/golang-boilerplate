package validator

import "testing"

type signup struct {
	FullName string `json:"full_name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `validate:"min=18"`
}

func TestValidateStruct(t *testing.T) {
	if errs := ValidateStruct(signup{FullName: "A", Email: "a@example.com", Age: 30}); errs != nil {
		t.Fatalf("valid struct: %v", errs)
	}

	errs := ValidateStruct(signup{Email: "nope", Age: 3})
	got := map[string]string{}
	for _, e := range errs {
		got[e.Field] = e.Message
	}
	want := map[string]string{
		"full_name": "full_name is required",
		"email":     "email must be a valid email address",
		"Age":       "Age must be at least 18 characters",
	}
	for field, msg := range want {
		if got[field] != msg {
			t.Errorf("%s: got %q, want %q", field, got[field], msg)
		}
	}
}

func TestValidateStructNonStruct(t *testing.T) {
	if errs := ValidateStruct(42); len(errs) != 1 {
		t.Errorf("non-struct input: %v", errs)
	}
}
