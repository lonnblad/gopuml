package internal_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lonnblad/gopuml/cmd/gopuml/internal"
)

func Test_RunRootCommand(t *testing.T) {
	const (
		expectedOutput = "Compiles Plant UML files\n\n"
	)

	cmd := internal.CreateRootCmd()

	var stdout, stderr bytes.Buffer

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Empty(t, stderr.String())
	assert.Equal(t, expectedOutput, stdout.String())
}
