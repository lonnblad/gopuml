package internal_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lonnblad/gopuml/cmd/gopuml/internal"
	"github.com/lonnblad/gopuml/example"
)

func Test_RunBuildCommand(t *testing.T) {
	testcases := createBuildCmdTestcases()

	for _, tc := range testcases {
		t.Run(tc.name("stdin"), tc.testStdin)
		t.Run(tc.name("args"), tc.testArgs)
	}
}

type buildCmdTestcase struct {
	format         string
	style          string
	expectedOutput string
}

func createBuildCmdTestcases() []buildCmdTestcase {
	return []buildCmdTestcase{
		{
			format: "png", style: "file",
			expectedOutput: example.PNGFile(),
		},
		{
			format: "svg", style: "file",
			expectedOutput: example.SVGFile(),
		},
		{
			format: "txt", style: "file",
			expectedOutput: example.TXTFile(),
		},
		{
			format: "png", style: "link",
			expectedOutput: example.PNGLink() + "\n",
		},
		{
			format: "svg", style: "link",
			expectedOutput: example.SVGLink() + "\n",
		},
		{
			format: "txt", style: "link",
			expectedOutput: example.TXTLink() + "\n",
		},
		{
			format: "png", style: "out",
			expectedOutput: example.PNGFile(),
		},
		{
			format: "svg", style: "out",
			expectedOutput: example.SVGFile(),
		},
		{
			format: "txt", style: "out",
			expectedOutput: example.TXTFile(),
		},
	}
}

func (tc buildCmdTestcase) args(args ...string) []string {
	return append([]string{"-f", tc.format, "--style", tc.style}, args...)
}

func (tc buildCmdTestcase) name(suffix string) string {
	return tc.format + "/" + tc.style + "/" + suffix
}

func (tc buildCmdTestcase) testStdin(t *testing.T) {
	t.Parallel()

	if tc.style == "file" {
		t.Skipf("stdin doesn't support style: [%s]", tc.style)
	}

	cmd := internal.CreateBuildCmd()
	cmd.SetArgs(tc.args())

	stdin := bytes.NewBufferString(example.PUML())
	cmd.SetIn(stdin)

	tc.executeAndValidate(t, t.TempDir(), cmd)
}

func (tc buildCmdTestcase) testArgs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inputFile := dir + "/" + "example.puml"

	err := os.WriteFile(inputFile, []byte(example.PUML()), 0666)
	require.Nil(t, err)

	cmd := internal.CreateBuildCmd()
	cmd.SetArgs(tc.args(inputFile))

	tc.executeAndValidate(t, t.TempDir(), cmd)
}

func (tc buildCmdTestcase) executeAndValidate(t *testing.T, tempDir string, cmd cobra.Command) {
	var stdout, stderr bytes.Buffer

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Empty(t, stderr.String())

	if tc.style != "file" {
		assert.Equal(t, tc.expectedOutput, stdout.String())
		return
	}

	outputFile := tempDir + "/" + "example." + tc.format
	actualOutput, err := os.ReadFile(outputFile)
	require.Nil(t, err)
	assert.Equal(t, tc.expectedOutput, string(actualOutput))
}
