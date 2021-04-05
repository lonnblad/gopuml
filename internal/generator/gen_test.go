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
	var expectedFile generator.File

	expectedFile.UpdatedAt = time.Now()
	expectedFile.Filename = "example.puml"
	expectedFile.Filepath = "<path>/" + expectedFile.Filename
	expectedFile.Raw = []byte(example.PUML())
	expectedFile.Encoded = []byte(example.EncodedPUML)

	gen := generator.New()

	err := gen.PutFile(expectedFile.Filepath, expectedFile.Raw)
	require.Nil(t, err)

	fs := gen.GetFiles()
	require.Len(t, fs, 1)

	f := fs[0]
	assert.Equal(t, expectedFile.Filepath, f.Filepath)
	assert.Equal(t, expectedFile.Filename, f.Filename)
	assert.WithinDuration(t, expectedFile.UpdatedAt, f.UpdatedAt, time.Millisecond)
	assert.Equal(t, expectedFile.Raw, f.Raw)
	assert.Equal(t, expectedFile.Encoded, f.Encoded)

	err = gen.PutFile(expectedFile.Filepath, expectedFile.Raw)
	require.Nil(t, err)

	fsTwo := gen.GetFiles()
	require.Len(t, fsTwo, 1)
	assert.Equal(t, fs, fsTwo)
}

func Test_Subs(t *testing.T) {
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

	err = gen.PutFile(expectedFile.Filepath, expectedFile.Raw)
	require.Nil(t, err)

	for n := 1; n <= 3; n++ {
		select {
		case f := <-testFileChan:
			require.True(t, n < 3)

			assert.Equal(t, expectedFile.Filepath, f.Filepath)
			assert.Equal(t, expectedFile.Filename, f.Filename)
			assert.WithinDuration(t, expectedFile.UpdatedAt, f.UpdatedAt, time.Millisecond)
			assert.Equal(t, expectedFile.Raw, f.Raw)
			assert.Equal(t, expectedFile.Encoded, f.Encoded)
		case <-time.After(time.Millisecond):
			assert.True(t, n == 3)
		}
	}
}
