package services

import (
	"pethelp-backend/internal/core/domain"
	"pethelp-backend/internal/core/ports"
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/nyaruka/phonenumbers"
)

type SpecialistValidatorImpl struct {
	validator *validator.Validate
}

var _ ports.SpecialistValidator = (*SpecialistValidatorImpl)(nil)

// Custom validation function for E.123 phone number
func isValidE123(fl validator.FieldLevel) bool {
	e123Regex := regexp.MustCompile(`^\+\d{1,3}(?:[()\s-]*\d+)*$`)
	digNumber := phonenumbers.NormalizeDigitsOnly(fl.Field().String())
	num, err := phonenumbers.Parse(digNumber, "UA")
	if err != nil {
		return false
	}
	if !phonenumbers.IsValidNumber(num) {
		return false
	}

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

func (sv *SpecialistValidatorImpl) ValidateRegistrationReq(data domain.RegistrationRequest) []domain.FieldError {
	var validationErrors []domain.FieldError

	if err := sv.validator.Struct(data); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
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
				fe.Message = "Phone must be compatible with E.164 and E.123 formats (e.g., +38(XXX)XXX-XX-XX and contain at least 13 digits."
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

	passwordErrors := validatePasswordComplexity(data.Password, "Password")
	if len(passwordErrors) > 0 {
		validationErrors = append(validationErrors, passwordErrors...)
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}

// ValidateChangePassword checks the change password request.
func (sv *SpecialistValidatorImpl) ValidateChangePasswordReq(reqData domain.ChangePassReq) []domain.FieldError {
	var validationErrors []domain.FieldError

	if err := sv.validator.Struct(reqData); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			validationErrors = append(validationErrors, domain.FieldError{Field: "general", Message: "Invalid validation object"})
			return validationErrors
		}

		for _, err := range err.(validator.ValidationErrors) {
			var fe domain.FieldError
			fe.Field = err.Field()
			switch err.Field() {
			case "NewPass":
				if err.Tag() == "necsfield" {
					fe.Message = "The new password must be different from the current password (case-sensitive)."
					break
				}
				fe.Message = err.Error()
			case "CurrentPass":
				fe.Message = "The current password is required."
			}

			validationErrors = append(validationErrors, fe)
		}
	}

	passwordErrors := validatePasswordComplexity(reqData.NewPass, "NewPass")
	if len(passwordErrors) > 0 {
		validationErrors = append(validationErrors, passwordErrors...)
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}

// ValidateSpecialistProfileUpdateReq checks the specialist profile update request.
func (sv *SpecialistValidatorImpl) ValidateSpecialistProfileUpdateReq(reqData domain.SpecialistProfUpdateReq) []domain.FieldError {
	var validationErrors []domain.FieldError

	if err := sv.validator.Struct(reqData); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			validationErrors = append(validationErrors, domain.FieldError{Field: "general", Message: "Invalid validation object"})
			return validationErrors
		}

		for _, err := range err.(validator.ValidationErrors) {
			var fe domain.FieldError
			fe.Field = err.Field()
			switch err.Field() {
			case "Name":
				fe.Message = "Invalid name. It must be 2-100 characters and contain only letters, spaces, hyphens, or apostrophes."
			case "FamilyName":
				fe.Message = "Invalid family name. It must be 2-100 characters and contain only letters, spaces, hyphens, or apostrophes."
			case "Phone":
				fe.Message = "Phone must be compatible with E.164 and E.123 formats (e.g., +38(XXX)XXX-XX-XX) and contain at least 13 digits."
			case "Experience":
				fe.Message = "Experience years must be a non-negative number."
			case "Bio":
				fe.Message = "Bio cannot exceed 1000 characters."
			default:
				fe.Message = err.Error() // Fallback for unhandled fields
			}

			validationErrors = append(validationErrors, fe)
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}

func validatePasswordComplexity(password, fieldName string) []domain.FieldError {

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
		errors = append(errors, domain.FieldError{Field: fieldName, Message: domain.ErrNoUppercase.Error()})
	}
	if !hasLower {
		errors = append(errors, domain.FieldError{Field: fieldName, Message: domain.ErrNoLowercase.Error()})
	}
	if !hasNumber {
		errors = append(errors, domain.FieldError{Field: fieldName, Message: domain.ErrNoNumber.Error()})
	}
	if !hasSpecial {
		errors = append(errors, domain.FieldError{Field: fieldName, Message: domain.ErrNoSpecialChar.Error()})
	}

	return errors
}
