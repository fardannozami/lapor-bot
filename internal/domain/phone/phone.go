package phone

import (
	"errors"
	"regexp"
	"strings"
)

var onlyDigit = regexp.MustCompile(`^\d+$`)

func Normalize(raw string) (string, error) {
	p := strings.TrimSpace(raw)
	p = strings.ReplaceAll(p, "-", "")
	p = strings.ReplaceAll(p, "+", "")
	p = strings.ReplaceAll(p, " ", "")

	if p == "" {
		return "", errors.New("phone number is required")
	}

	if !onlyDigit.MatchString(p) {
		return "", errors.New("phone must contain only digit")
	}

	if len(p) < 9 || len(p) > 15 {
		return "", errors.New("phone length must between 9 and 15 char")
	}

	return p, nil
}
