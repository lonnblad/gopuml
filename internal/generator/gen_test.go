package generator_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lonnblad/gopuml/example"
	"github.com/lonnblad/gopuml/internal/generator"
)

func Test_Generator(t *testing.T) {
	var expectedFile generator.File

	expectedFile.UpdatedAt = time.Now()
	expectedFile.Filename = "example.puml"
	expectedFile.Filepath = "<path>/" + expectedFile.Filename
	expectedFile.Raw = []byte(example.PUML())
	expectedFile.Encoded = []byte(example.EncodedPUML)

	gen := generator.New()

	testFileChan := make(chan generator.File, 3)

	idOne, cOne := gen.RegisterSub()
	defer gen.DeregisterSub(idOne)

	go func() {
		for f := range cOne {
			testFileChan <- f
		}
	}()

	idTwo, cTwo := gen.RegisterSub()
	defer gen.DeregisterSub(idTwo)

	go func() {
		for f := range cTwo {
			testFileChan <- f
		}
	}()

	err := gen.PutFile(expectedFile.Filepath, expectedFile.Raw)
	require.Nil(t, err)

	fs := gen.GetFiles()
	require.Len(t, fs, 1)

	actualFile := fs[0]
	equalFile(t, expectedFile, actualFile)

	err = gen.PutFile(expectedFile.Filepath, expectedFile.Raw)
	require.Nil(t, err)

	for n := 1; n <= 3; n++ {
		select {
		case actualFile := <-testFileChan:
			require.True(t, n < 3)
			equalFile(t, expectedFile, actualFile)
		case <-time.After(time.Millisecond):
			assert.True(t, n == 3)
		}
	}

	fsTwo := gen.GetFiles()
	require.Len(t, fsTwo, 1)
	assert.Equal(t, fs, fsTwo)
}

func equalFile(t *testing.T, expected, actual generator.File) {
	assert.Equal(t, expected.Filepath, actual.Filepath)
	assert.Equal(t, expected.Filename, actual.Filename)
	assert.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, time.Millisecond)
	assert.Equal(t, expected.Raw, actual.Raw)
	assert.Equal(t, expected.Encoded, actual.Encoded)
}
