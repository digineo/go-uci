package uci

import (
	"fmt"
)

// ErrConfigAlreadyLoaded is returned by LoadConfig, if the given config
// name is already present.
type ErrConfigAlreadyLoaded struct {
	Name string
}

func (err ErrConfigAlreadyLoaded) Error() string {
	return fmt.Sprintf("%s already loaded", err.Name)
}

// ErrUnknownOptionType is returned when trying to parse an invalid OptionType.
type ErrUnknownOptionType struct {
	Type string
}

func (err ErrUnknownOptionType) Error() string {
	return fmt.Sprintf("Unknown Option type %s", err.Type)
}

// ErrSectionTypeMismatch is returned by AddSection if the section-to-add
// already exists with a different type.
type ErrSectionTypeMismatch struct {
	Config, Section string // name
	ExistingType    string
	NewType         string
}

func (err ErrSectionTypeMismatch) Error() string {
	return fmt.Sprintf("type mismatch for %s.%s, got %s, want %s",
		err.Config, err.Section, err.ExistingType, err.NewType)
}

type ParseError struct {
	errstr string
	token  token
}

func (err ParseError) Error() string {
	if err.token.typ != scanToken(0) { // check if we got a valid token, or if it is a generic parse error
		return fmt.Sprintf("parse error: %s, token: %s", err.errstr, err.token.String())
	}
	return fmt.Sprintf("parse error: %s", err.errstr)
}

// ErrSectionNotFound is returned by Get
type ErrSectionNotFound struct {
	Section string
}

func (err ErrSectionNotFound) Error() string {
	return fmt.Sprintf("section %s not found", err.Section)
}
