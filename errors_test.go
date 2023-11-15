package uci

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrConfigAlreadyLoaded(t *testing.T) {
	assert := assert.New(t)

	err := ErrConfigAlreadyLoaded{"foo"}
	assert.Equal(err.Error(), "foo already loaded")

	assert.False(IsConfigAlreadyLoaded(nil))
	assert.True(IsConfigAlreadyLoaded(&err))
}

func TestErrSectionTypeMismatch(t *testing.T) {
	assert := assert.New(t)

	err := ErrSectionTypeMismatch{"foo", "bar", "interface", "radio"}
	assert.Equal(err.Error(), "type mismatch for foo.bar, got interface, want radio")

	assert.False(IsSectionTypeMismatch(nil))
	assert.True(IsSectionTypeMismatch(&err))
}

func TestParseError(t *testing.T) {
	assert := assert.New(t)

	perr := ParseError{errstr: "expected foo"}
	err := fmt.Errorf("parse errstr: %w", perr)
	assert.Contains(err.Error(), "parse errstr: expected foo")

	assert.False(IsParseError(nil))
	assert.True(IsParseError(err))
}
