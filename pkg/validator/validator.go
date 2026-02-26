package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var (
	uppercaseRegex   = regexp.MustCompile(`[A-Z]`)
	lowercaseRegex   = regexp.MustCompile(`[a-z]`)
	digitRegex       = regexp.MustCompile(`[0-9]`)
	specialCharRegex = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`)
)

// PasswordStrength validates password strength requirements:
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one digit
// - At least one special character
func PasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if !uppercaseRegex.MatchString(password) {
		return false
	}
	if !lowercaseRegex.MatchString(password) {
		return false
	}
	if !digitRegex.MatchString(password) {
		return false
	}
	if !specialCharRegex.MatchString(password) {
		return false
	}

	return true
}

// RegisterCustomValidators registers all custom validators with the given validator instance
func RegisterCustomValidators(v *validator.Validate) error {
	return v.RegisterValidation("password_strength", PasswordStrength)
}
