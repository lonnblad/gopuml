package internal

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/lonnblad/gopuml/internal/generator"
)

const (
	defaultPort = "8080"

	flagPort, flagShortPort = "port", "p"
)

type serveOptions struct {
	Port string
}

const flagUsagePort = `the port to use to serve the HTML page
 `

// CreateServeCmd creates the serve subcommand.
// The command will run a webserver which renders the supplied Plant UML files as a static HTML page.
// The command uses a file watcher to keep track of any modifications to the supplied files.
// The HTML page execute HEAD requests to check for new updates using long-polling and the If-Modified-Since header.
// When modifications are found, the server will answer the HEAD request with a 200 OK.
func CreateServeCmd() cobra.Command {
	opts := serveOptions{
		Port: defaultPort,
	}

	serveCmd := cobra.Command{
		Use:   "serve",
		Short: "Starts a web server which serves compiled UML files on a static HTML page.",
		Long: `Starts a web server which serves compiled UML files.
On modifications to the files, the HTML page will reload.`,
		RunE: serveCmdRunFunc(&opts),
	}

	serveCmd.Flags().StringVarP(&opts.Port, flagPort, flagShortPort, opts.Port, flagUsagePort)

	return serveCmd
}

func serveCmdRunFunc(opts *serveOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		generator := generator.New()

		fileWatcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		defer fileWatcher.Close()

		go eventHandler(cmd, fileWatcher, generator)

		if err = readAllFiles(fileWatcher, generator, args); err != nil {
			return err
		}

		if err = runServer(cmd, opts.Port, generator); err != nil {
			return err
		}

		return nil
	}
}

func eventHandler(cmd *cobra.Command, watcher *fsnotify.Watcher, gen *generator.Generator) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				path := event.Name

				fmt.Fprintln(cmd.OutOrStdout(), "modified file:", path)

				content, err := os.ReadFile(path)
				if err != nil {
					cmd.PrintErrln(err)
					return
				}

				if err = gen.PutFile(path, content); err != nil {
					cmd.PrintErrln(err)
					return
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			cmd.PrintErrln(err)
		}
	}
}

func readAllFiles(fileWatcher *fsnotify.Watcher, gen *generator.Generator, args []string) error {
	filepaths, err := findAbsolutePaths(args)
	if err != nil {
		return err
	}

	for _, path := range filepaths {
		if err = fileWatcher.Add(path); err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if err = gen.PutFile(path, content); err != nil {
			return err
		}
	}

	return nil
}

func runServer(cmd *cobra.Command, port string, gen *generator.Generator) error {
	server := &http.Server{Addr: ":" + port, Handler: handler(gen)}

	fmt.Fprintln(cmd.OutOrStdout(), "Server started")
	fmt.Fprintf(cmd.OutOrStdout(), "  http://localhost:%s\n\n", port)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

const (
	contentType = "Content-Type"
	mimeHTML    = "text/html"
)

func handler(gen *generator.Generator) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "HEAD" {
			handleHEAD(gen, w, req)
			return
		}

		w.Header().Set(contentType, mimeHTML)

		content, err := buildHTML(gen.GetFiles())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("")) // nolint: errcheck

			return
		}

		w.Write(content) // nolint: errcheck
	})

	return mux
}

func handleHEAD(gen *generator.Generator, w http.ResponseWriter, req *http.Request) {
	content := []byte("")

	etag := req.Header.Get("If-Modified-Since")
	if etag == "" {
		w.Write(content) // nolint: errcheck
		return
	}

	since, err := time.Parse(time.RFC1123, etag)
	if err != nil {
		w.Write(content) // nolint: errcheck
		return
	}

	id, contentChan := gen.RegisterSub()
	defer gen.DeregisterSub(id)

	for _, file := range gen.GetFiles() {
		if file.UpdatedAt.After(since) {
			w.Write(content) // nolint: errcheck
			return
		}
	}

	const longPollingTimeout = 60 * time.Second

	select {
	case <-contentChan:
	case <-time.After(longPollingTimeout):
		w.WriteHeader(http.StatusNotModified)
	}

	w.Write(content) // nolint: errcheck
}

func buildHTML(files []generator.File) (_ []byte, err error) {
	generator, err := template.New("html_page").Parse(htmlPageTemplate)
	if err != nil {
		err = fmt.Errorf("failed to parse HTML page template: %w", err)
		return
	}

	type file struct {
		Filename string
		PngLink  string
		SvgLink  string
	}

	var templateInfo = struct{ Files []file }{Files: make([]file, len(files))}

	for idx, f := range files {
		templateInfo.Files[idx].Filename = f.Filename
		templateInfo.Files[idx].PngLink = createLink(defaultServer, formatPNG, f.Encoded)
		templateInfo.Files[idx].SvgLink = createLink(defaultServer, formatSVG, f.Encoded)
	}

	var buffer bytes.Buffer
	if err = generator.Execute(&buffer, templateInfo); err != nil {
		err = fmt.Errorf("failed to execute generator: %w", err)
		return
	}

	return buffer.Bytes(), nil
}

const htmlPageTemplate = `<!DOCTYPE html>
<html lang=en>
<head>
  <title>gopuml</title>
  <meta name='generator' content='github.com/lonnblad/gopuml'>
</head>
<body onload="javascript:checkReload();" style="width:100vw;height:100vh;background-color:lightgrey;">
  <div style="margin: 0px 20px;width:100%">
    {{range .Files}}
    <h2>{{.Filename}}</h2>

    <h3>.png</h3>
		Static <a href="{{.PngLink}}">.png link</a> from plantuml.com.
    <p>
      <img style="object-fit:contain;" src="{{.PngLink}}" alt=".png" />
    </p>

    <h3>.svg</h3>
		Static <a href="{{.SvgLink}}">.svg link</a> from plantuml.com.
    <p>
      <img style="object-fit:contain;" src="{{.SvgLink}}" alt=".svg" />
    </p>
    {{end}}
  </div>
  <script>
    function checkReload() {
			const timestamp = new Date().toUTCString();

      fetch('/', {method: 'HEAD', headers: {'If-Modified-Since': timestamp}})
        .then(response => {
					if (response.status === 200) {
						location.reload(true)
					} else {
						checkReload()
					}
				})
				.catch(err => console.log(err));
    }
  </script>
</body>
</html>`
