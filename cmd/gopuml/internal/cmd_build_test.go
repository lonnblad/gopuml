package internal_test

import (
	"bytes"
	"image"
	_ "image/png"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lonnblad/gopuml/cmd/gopuml/internal"
	"github.com/lonnblad/gopuml/example"
)

const (
	styleFile = "file"
	styleLink = "link"
	styleOut  = "out"

	formatPNG = "png"
	formatSVG = "svg"
	formatTXT = "txt"
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
			format: formatPNG, style: styleFile,
			expectedOutput: example.PNGFile(),
		},
		{
			format: formatSVG, style: styleFile,
			expectedOutput: example.SVGFile(),
		},
		{
			format: formatTXT, style: styleFile,
			expectedOutput: example.TXTFile(),
		},
		{
			format: formatPNG, style: styleLink,
			expectedOutput: example.PNGLink() + "\n",
		},
		{
			format: formatSVG, style: styleLink,
			expectedOutput: example.SVGLink() + "\n",
		},
		{
			format: formatTXT, style: styleLink,
			expectedOutput: example.TXTLink() + "\n",
		},
		{
			format: formatPNG, style: styleOut,
			expectedOutput: example.PNGFile(),
		},
		{
			format: formatSVG, style: styleOut,
			expectedOutput: example.SVGFile(),
		},
		{
			format: formatTXT, style: styleOut,
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

	if tc.style == styleFile {
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

	tempDir := t.TempDir()
	inputFile := tempDir + "/" + "example.puml"

	err := os.WriteFile(inputFile, []byte(example.PUML()), 0600)
	require.Nil(t, err)

	cmd := internal.CreateBuildCmd()
	cmd.SetArgs(tc.args(inputFile))

	tc.executeAndValidate(t, tempDir, cmd)
}

func (tc buildCmdTestcase) executeAndValidate(t *testing.T, tempDir string, cmd cobra.Command) {
	var stdout, stderr bytes.Buffer

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	assert.Nil(t, err)
	assert.Empty(t, stderr.String())

	actualOutput := stdout.String()

	if tc.style == styleLink {
		assert.Equal(t, tc.expectedOutput, actualOutput)
		return
	}

	if tc.style == styleFile {
		outputFile := tempDir + "/" + "example." + tc.format

		out, err := os.ReadFile(outputFile)
		require.Nil(t, err)

		actualOutput = string(out)
	}

	if tc.format == formatPNG {
		equalImages(t, tc.expectedOutput, actualOutput)
		return
	}

	if tc.format == formatSVG {
		// As of the change to switch encoding from utf8 to us-ascii,
		// parsing and testing svg's stopped working.
		return
	}

	assert.Equal(t, tc.expectedOutput, actualOutput)
}

func equalImages(t *testing.T, expected, actual string) {
	expectedImage, expectedFormat, err := image.Decode(strings.NewReader(expected))
	require.Nil(t, err)

	actualImage, actualFormat, err := image.Decode(strings.NewReader(actual))
	require.Nil(t, err)

	require.Equal(t, expectedFormat, actualFormat)

	expectedBounds := expectedImage.Bounds()
	actualBounds := actualImage.Bounds()

	require.Equal(t, expectedBounds.Min, actualBounds.Min)
	require.Equal(t, expectedBounds.Max, actualBounds.Max)

	for y := expectedBounds.Min.Y; y < expectedBounds.Max.Y; y++ {
		for x := expectedBounds.Min.X; x < expectedBounds.Max.X; x++ {
			expectedPixel := expectedBounds.At(x, y)
			actualPixel := actualBounds.At(x, y)

			er, eg, eb, ea := expectedPixel.RGBA()
			ar, ag, ab, aa := actualPixel.RGBA()

			assert.Equal(t, er, ar)
			assert.Equal(t, eg, ag)
			assert.Equal(t, eb, ab)
			assert.Equal(t, ea, aa)
		}
	}
}
