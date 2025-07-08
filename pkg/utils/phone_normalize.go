package utils

import (
	"fmt"

	"github.com/nyaruka/phonenumbers"
)

func NormalizePhoneNumber(rawPhone string) (string, error) {
	//use ukraine country code by default
	num, err := phonenumbers.Parse(rawPhone, "UA")
	if err != nil {
		return "", fmt.Errorf("failed to parse phone number: %w", err)
	}
	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("invalid phone number")
	}
	return phonenumbers.Format(num, phonenumbers.E164), nil
}
