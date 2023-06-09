package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lonnblad/gopuml"
)

const (
	defaultFormat = formatSVG
	defaultStyle  = styleFile
	defaultServer = "https://www.plantuml.com/plantuml"

	flagStyle                   = "style"
	flagFormat, flagShortFormat = "format", "f"
	flagServer                  = "server"

	styleFile = "file"
	styleLink = "link"
	styleOut  = "out"

	formatPNG = "png"
	formatSVG = "svg"
	formatTXT = "txt"
)

type buildOptions struct {
	Server string
	Style  string
	Format string
}

const flagUsageStyle = `the style in which to compile the files

supported styles are:
  ` + styleFile + `  will write the formatted content to a file
  ` + styleLink + `  will write a link to the formatted content to stdout
  ` + styleOut + `   will write the formatted content to stdout
 `

const flagUsageFormat = `the format of the compiled files

supported formatters are:
  ` + formatPNG + `  will format the content as .png
  ` + formatSVG + `  will format the content as .svg
  ` + formatTXT + `  will format the content as .txt
 `

const flagUsageServer = `the Server URL to use when the style used is link,

the provided server need to support links formatted like:
  "<server_url>/<format>/<plant_uml_text_encoding>"
 `

// CreateBuildCmd creates the build subcommand.
func CreateBuildCmd() cobra.Command {
	opts := buildOptions{
		Server: defaultServer,
		Style:  defaultStyle,
		Format: defaultFormat,
	}

	buildCmd := cobra.Command{
		Use:   "build [plant UML files]",
		Short: "Compiles Plant UML files",
		Example: `  gopuml build example.puml
  gopuml build -f png --style link example.puml`,
		RunE: buildCmdRunFunc(&opts),
	}

	buildCmd.Flags().StringVarP(&opts.Format, flagFormat, flagShortFormat, opts.Format, flagUsageFormat)
	buildCmd.Flags().StringVar(&opts.Server, flagServer, opts.Server, flagUsageServer)
	buildCmd.Flags().StringVar(&opts.Style, flagStyle, opts.Style, flagUsageStyle)

	return buildCmd
}

func buildCmdRunFunc(opts *buildOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return buildFromStdIn(opts, cmd)
		}

		return buildFromArgs(opts, cmd, args)
	}
}

func buildFromStdIn(opts *buildOptions, cmd *cobra.Command) error {
	content, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return err
	}

	if content, err = compressAndEncode(content); err != nil {
		return err
	}

	if err = opts.writeOutput(cmd.OutOrStdout(), content); err != nil {
		return err
	}

	return nil
}

func buildFromArgs(opts *buildOptions, cmd *cobra.Command, args []string) error {
	filepaths, err := findAbsolutePaths(args)
	if err != nil {
		return err
	}

	for _, file := range filepaths {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		if content, err = compressAndEncode(content); err != nil {
			return err
		}

		output := cmd.OutOrStdout()

		if opts.Style == styleFile {
			outputFilename := strings.TrimSuffix(file, filepath.Ext(file))
			outputFilename = fmt.Sprintf("%s.%s", outputFilename, opts.Format)

			var f *os.File

			const readWriteMode = 0600
			if f, err = os.OpenFile(outputFilename, os.O_RDWR|os.O_CREATE, readWriteMode); err != nil {
				return fmt.Errorf("couldn't open file: %w", err)
			}

			defer f.Close()

			output = f
		}

		if err = opts.writeOutput(output, content); err != nil {
			return err
		}
	}

	return nil
}

func (opts buildOptions) writeOutput(out io.Writer, content []byte) error {
	switch opts.Style {
	case styleLink:
		link := createLink(opts.Server, opts.Format, content)
		fmt.Fprintln(out, link)
	case styleFile, styleOut:
		link := createLink(opts.Server, opts.Format, content)

		response, err := http.Get(link) // nolint: gosec
		if err != nil {
			return fmt.Errorf("can't fetch output: %w", err)
		}

		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("wrong status code %s, when fetching link: %s", response.Status, link)
		}

		if _, err = io.Copy(out, response.Body); err != nil {
			return fmt.Errorf("couldn't write to output: %w", err)
		}
	}

	return nil
}

func findAbsolutePaths(args []string) (_ []string, err error) {
	filepaths := make([]string, 0, len(args))
	uniqueFileMap := make(map[string]bool)

	for _, filename := range args {
		var absolutePath string

		if absolutePath, err = filepath.Abs(filename); err != nil {
			err = fmt.Errorf("unable to resolve filename: [%s]: %w", filename, err)
			return
		}

		if _, ok := uniqueFileMap[absolutePath]; !ok {
			filepaths = append(filepaths, absolutePath)
			uniqueFileMap[absolutePath] = true
		}
	}

	return filepaths, nil
}

func compressAndEncode(data []byte) (_ []byte, err error) {
	if data, err = gopuml.Deflate(data); err != nil {
		err = fmt.Errorf("couldn't compress the data: %w", err)
		return
	}

	return gopuml.Encode(data), nil
}

func createLink(server, format string, data []byte) string {
	return fmt.Sprintf("%s/%s/%s", server, format, string(data))
}
