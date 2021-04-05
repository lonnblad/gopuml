package generator_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lonnblad/gopuml/example"
	"github.com/lonnblad/gopuml/internal/generator"
)

func Test_Basic(t *testing.T) {
	testStartedAt := time.Now()

	const (
		expectedFilename = "example.puml"
		expectedFilepath = "<path>/" + expectedFilename
	)

	expectedContent := example.PUML()
	expectedEncoded := example.EncodedPUML

	gen := generator.New()

	err := gen.PutFile(expectedFilepath, []byte(expectedContent))
	require.Nil(t, err)

	fs := gen.GetFiles()
	require.Len(t, fs, 1)

	assert.Equal(t, expectedFilepath, fs[0].Filepath)
	assert.Equal(t, expectedFilename, fs[0].Filename)
	assert.WithinDuration(t, testStartedAt, fs[0].UpdatedAt, time.Millisecond)
	assert.Equal(t, expectedContent, string(fs[0].Raw))
	assert.Equal(t, expectedEncoded, string(fs[0].Encoded))

	err = gen.PutFile(expectedFilepath, []byte(expectedContent))
	require.Nil(t, err)

	fsTwo := gen.GetFiles()
	require.Len(t, fsTwo, 1)

	assert.Equal(t, fs, fsTwo)
}
