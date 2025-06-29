package services

import (
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	"regexp"
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
	v.RegisterValidation("custom_name", isValidName)

	return &SpecialistValidatorImpl{validator: v}
}

func (sv *SpecialistValidatorImpl) Validate(data *domain.RegistrationRequest) []domain.FieldError {
	var validationErrors []domain.FieldError

	if err := sv.validator.Struct(data); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			// This is an internal error, should not happen with a valid struct.
			validationErrors = append(validationErrors, domain.FieldError{Field: "general", Message: "Invalid validation object"})
			return validationErrors
		}

		for _, err := range err.(validator.ValidationErrors) {
			var fe domain.FieldError
			fe.Field = err.Field()
			switch err.Field() {
			case "Name":
				fe.Message = "Invalid name. It must be 2-100 characters and contain only letters, spaces, hyphens, or apostrophes."
			case "Phone":
				fe.Message = "Phone must be in E.123 format (e.g., +3(XXX)XXX-XX-XX) and contain at least 13 digits."
			case "Email":
				fe.Message = "Invalid email format."
			case "Password":
				fe.Message = "Password must be at least 12 characters."
			case "PasswordConfirmation":
				fe.Message = "Passwords do not match."
			}
			validationErrors = append(validationErrors, fe)
		}
	}

	passwordErrors := validatePasswordComplexity(data.Password)
	if len(passwordErrors) > 0 {
		validationErrors = append(validationErrors, passwordErrors...)
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}

func validatePasswordComplexity(password string) []domain.FieldError {

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	var errors []domain.FieldError

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
	}

	if !hasUpper {
		errors = append(errors, domain.FieldError{Field: "Password", Message: domain.ErrNoUppercase.Error()})
	}
	if !hasLower {
		errors = append(errors, domain.FieldError{Field: "Password", Message: domain.ErrNoLowercase.Error()})
	}
	if !hasNumber {
		errors = append(errors, domain.FieldError{Field: "Password", Message: domain.ErrNoNumber.Error()})
	}
	if !hasSpecial {
		errors = append(errors, domain.FieldError{Field: "Password", Message: domain.ErrNoSpecialChar.Error()})
	}

	return errors
}
