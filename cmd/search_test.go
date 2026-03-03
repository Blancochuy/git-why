package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHighlightTerm(t *testing.T) {
	plainOutput = false
	useRegex = false

	result := highlightTerm("fix: bug in authentication", "authentication")
	assert.Contains(t, result, "authentication")
	assert.Contains(t, result, "\033[33;1m")
}

func TestHighlightTermWithRegex(t *testing.T) {
	plainOutput = false
	useRegex = true

	result := highlightTerm("fix: security vulnerability in auth", "sec.*vuln")
	assert.Contains(t, result, "\033[33;1m")
}

func TestHighlightTermPlain(t *testing.T) {
	plainOutput = true
	result := highlightTerm("fix: security issue", "security")
	assert.Equal(t, "fix: security issue", result)
}
