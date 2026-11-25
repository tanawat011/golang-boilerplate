// Package validator validates request structs and formats the errors.
package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a single invalid field.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	// Report fields by their JSON name ("full_name"), not the Go name.
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			return f.Name
		}
		return name
	})
	return v
}

// ValidateStruct validates s and returns one entry per invalid field, or nil.
func ValidateStruct(s any) []ValidationError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var fieldErrs validator.ValidationErrors
	if !errors.As(err, &fieldErrs) {
		return []ValidationError{{Field: "", Message: err.Error()}}
	}

	out := make([]ValidationError, 0, len(fieldErrs))
	for _, fe := range fieldErrs {
		out = append(out, ValidationError{Field: fe.Field(), Message: message(fe)})
	}
	return out
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", fe.Field(), fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "alphanum":
		return fmt.Sprintf("%s may contain only letters and digits", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}
