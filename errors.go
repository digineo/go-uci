package uci

import "fmt"

// ErrConfigAlreadyLoaded is returned by LoadConfig, if the given config
// name is already present.
type ErrConfigAlreadyLoaded struct {
	Name string
}

func (err ErrConfigAlreadyLoaded) Error() string {
	return fmt.Sprintf("%s already loaded", err.Name)
}

// IsConfigAlreadyLoaded reports, whether err is of type ErrConfigAlredyLoaded.
func IsConfigAlreadyLoaded(err error) bool {
	if err == nil {
		return false
	}
	_, is := err.(*ErrConfigAlreadyLoaded)
	return is
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

// IsSectionTypeMismatch reports, whether err is of type ErrSectionTypeMismatch.
func IsSectionTypeMismatch(err error) bool {
	if err == nil {
		return false
	}
	_, is := err.(*ErrSectionTypeMismatch)
	return is
}

type ParseError string

func (err ParseError) Error() string {
	return fmt.Sprintf("parse error: %s", string(err))
}

func IsParseError(err error) bool {
	if err == nil {
		return false
	}
	_, is := err.(*ParseError)
	return is
}
