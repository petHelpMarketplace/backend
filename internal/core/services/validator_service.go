package services

import (
	"errors"
	"fmt"
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

type SpecialistValidatorImpl struct {
	validator *validator.Validate
}

var _ ports.SpecialistValidator = (*SpecialistValidatorImpl)(nil)

// Custom validation function for E.123 phone number
func isValidE123(fl validator.FieldLevel) bool {
	e123Regex := regexp.MustCompile(`^\+(?:\[\d{1,3}\]|\d{1,3})(?:[\s.-]?\d+)*$`)
	return e123Regex.MatchString(fl.Field().String())
}

// Custom validation function for allows Unicode letters, spaces, hyphens, and apostrophes.
func isValidName(fl validator.FieldLevel) bool {
	// `\p{L}` matches any kind of letter from any language.
	// `[\p{L}\s\-\']+` means one or more occurrences of (letter OR space OR hyphen OR apostrophe).
	nameRegex := regexp.MustCompile(`^[\p{L}\s\-\']+$`)
	return nameRegex.MatchString(fl.Field().String())
}

func NewCustomValidator() *SpecialistValidatorImpl {
	v := validator.New()
	v.RegisterValidation("e123", isValidE123)

	return &SpecialistValidatorImpl{validator: v}
}

func (sv *SpecialistValidatorImpl) Validate(data *domain.RegistrationRequest) error {
	sv.validator.RegisterValidation("e123", isValidE123)
	sv.validator.RegisterValidation("custom_name", isValidName)

	if err := sv.validator.Struct(data); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			case "Name", "FamilyName":
				errorMessages = append(errorMessages, fmt.Sprintf("%s must be at least 2 characters %s", err.Field(), err.Error()))
			case "Phone":
				errorMessages = append(errorMessages, "Phone must be in E.123 format (e.g., +38 (XXX) XXX-XX-XX)")
			case "Email":
				errorMessages = append(errorMessages, "Invalid email format")
			case "Password":
				errorMessages = append(errorMessages, "Password must be at least 12 characters")
			case "PasswordConfirmation":
				errorMessages = append(errorMessages, "Passwords do not match")
			}
		}
		return errors.New(strings.Join(errorMessages, "; "))
	}
	return isValidPassword(data.Password)
}

func isValidPassword(password string) error {

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}

		if hasUpper && hasLower && hasNumber && hasSpecial {
			return nil
		}
	}

	if !hasUpper {
		return domain.ErrNoUppercase
	}
	if !hasLower {
		return domain.ErrNoLowercase
	}
	if !hasNumber {
		return domain.ErrNoNumber
	}
	if !hasSpecial {
		return domain.ErrNoSpecialChar
	}

	return nil
}
