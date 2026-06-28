package game

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const MaxNameLength = 20

var (
	ErrEmptyName = errors.New("name cannot be empty")
	ErrNameLong  = errors.New("name must be 20 characters or fewer")
)

type Name struct {
	Display string
	Key     string
}

func NormalizeName(raw string) (Name, error) {
	display := strings.TrimSpace(raw)
	if display == "" {
		return Name{}, ErrEmptyName
	}
	if utf8.RuneCountInString(display) > MaxNameLength {
		return Name{}, ErrNameLong
	}

	key := strings.ToUpper(strings.ReplaceAll(display, " ", ""))
	if key == "" {
		return Name{}, ErrEmptyName
	}
	return Name{Display: display, Key: key}, nil
}
