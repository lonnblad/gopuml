package internal_test

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lonnblad/gopuml/cmd/gopuml/internal"
	"github.com/lonnblad/gopuml/example"
)

type buildCmdTestcase struct {
	name           string
	args           []string
	expectedOutput string
}

func Test_RunBuildCommand(t *testing.T) {
	testcases := []buildCmdTestcase{
		{
			name:           "png_link",
			args:           []string{"-f=png", "--style=link"},
			expectedOutput: example.PNGLink() + "\n",
		},
		{
			name:           "svg_link",
			args:           []string{"-f=svg", "--style=link"},
			expectedOutput: example.SVGLink() + "\n",
		},
		{
			name:           "txt_link",
			args:           []string{"-f=txt", "--style=link"},
			expectedOutput: example.TXTLink() + "\n",
		},
		{
			name:           "png_out",
			args:           []string{"-f=png", "--style=out"},
			expectedOutput: example.PNGFile(),
		},
		{
			name:           "svg_out",
			args:           []string{"-f=svg", "--style=out"},
			expectedOutput: example.SVGFile(),
		},
		{
			name:           "txt_out",
			args:           []string{"-f=txt", "--style=out"},
			expectedOutput: example.TXTFile(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name+"/stdin", tc.testStdIn)
		t.Run(tc.name+"/args", tc.testArgs)
	}
}

func (tc buildCmdTestcase) testStdIn(t *testing.T) {
	t.Parallel()

	cmd := internal.CreateBuildCmd()

	cmd.SetArgs(tc.args)

	stdin := bytes.NewBufferString(example.PUML())
	cmd.SetIn(stdin)

	var stdout, stderr bytes.Buffer

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Empty(t, stderr.String())
	assert.Equal(t, tc.expectedOutput, stdout.String())
}

func (tc buildCmdTestcase) testArgs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inputFile := dir + "/" + "example.puml"

	// nolint:gosec
	out, err := exec.Command("bash", "-c", "echo -n \""+example.PUML()+"\" > "+inputFile).CombinedOutput()
	require.Nilf(t, err, "out: %s", string(out))

	cmd := internal.CreateBuildCmd()

	args := append(tc.args, inputFile)
	cmd.SetArgs(args)

	var stdout, stderr bytes.Buffer

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err = cmd.Execute()
	assert.Nil(t, err)
	assert.Empty(t, stderr.String())
	assert.Equal(t, tc.expectedOutput, stdout.String())
}
