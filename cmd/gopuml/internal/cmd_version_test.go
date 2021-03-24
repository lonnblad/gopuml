package internal_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lonnblad/gopuml/cmd/gopuml/internal"
)

func Test_RunVersionCommand(t *testing.T) {
	const (
		version        = "test"
		expectedOutput = "gopuml version is: " + version + "\n"
	)

	cmd := internal.CreateVersionCmd(version)

	var stdout, stderr bytes.Buffer

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Empty(t, stderr.String())
	assert.Equal(t, expectedOutput, stdout.String())
}
