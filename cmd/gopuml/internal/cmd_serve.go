package internal

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
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
		var htmlCreator htmlCreator

		fileWatcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		defer fileWatcher.Close()

		go eventHandler(cmd, fileWatcher, &htmlCreator)

		if err = compileAllFiles(args, fileWatcher, &htmlCreator); err != nil {
			return err
		}

		if err = runServer(cmd, opts.Port, &htmlCreator); err != nil {
			return err
		}

		return nil
	}
}

func eventHandler(cmd *cobra.Command, watcher *fsnotify.Watcher, htmlCreator *htmlCreator) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				handleFileWriteNotification(cmd, htmlCreator, event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			cmd.Println("error:", err)
		}
	}
}

func handleFileWriteNotification(cmd *cobra.Command, htmlCreator *htmlCreator, path string) {
	cmd.Println("modified file:", path)

	newFile, err := readAndEncodeFile(path)
	if err != nil {
		cmd.PrintErrln(err)
	}

	oldFile := htmlCreator.getCompiledFile(path)

	if bytes.Equal(oldFile.compressed, newFile.compressed) {
		return
	}

	if err = htmlCreator.setCompiledFile(path, newFile); err != nil {
		cmd.PrintErrln(err)
	}
}

func compileAllFiles(args []string, fileWatcher *fsnotify.Watcher, htmlCreator *htmlCreator) error {
	filepaths, err := findAbsolutePaths(args)
	if err != nil {
		return err
	}

	for _, path := range filepaths {
		if err = fileWatcher.Add(path); err != nil {
			return err
		}

		file, err := readAndEncodeFile(path)
		if err != nil {
			return err
		}

		if err = htmlCreator.setCompiledFile(path, file); err != nil {
			return err
		}
	}

	return nil
}

func runServer(cmd *cobra.Command, port string, htmlCreator *htmlCreator) error {
	server := &http.Server{Addr: ":" + port, Handler: handler(htmlCreator)}

	cmd.Println("Server started")
	cmd.Printf("  http://localhost:%s\n\n", port)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

func readAndEncodeFile(path string) (_ compiledFile, err error) {
	file := compiledFile{Filename: filepath.Base(path)}

	var content []byte

	if content, err = os.ReadFile(path); err != nil {
		return
	}

	if content, err = compressAndEncode(content); err != nil {
		return
	}

	file.compressed = content

	file.PngLink = createLink(defaultServer, formatPNG, content)
	file.SvgLink = createLink(defaultServer, formatSVG, content)

	return file, nil
}

type compiledFile struct {
	Filename   string
	compressed []byte
	PngLink    string
	SvgLink    string
}

type htmlCreator struct {
	html      []byte
	updatedAt time.Time

	files map[string]compiledFile
	subs  map[int]chan bool

	noOfSubs int

	mutex sync.RWMutex
}

func (ch *htmlCreator) setCompiledFile(filename string, file compiledFile) error {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	if ch.files == nil {
		ch.files = make(map[string]compiledFile)
	}

	ch.files[filename] = file

	html, err := buildHTML(ch.files)
	if err != nil {
		return err
	}

	ch.html = html
	ch.updatedAt = time.Now()

	for _, sub := range ch.subs {
		sub <- true
	}

	return nil
}

func (ch *htmlCreator) getCompiledFile(filename string) compiledFile {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	return ch.files[filename]
}

func (ch *htmlCreator) getHTMLContent() ([]byte, time.Time) {
	ch.mutex.RLock()
	defer ch.mutex.RUnlock()

	return ch.html, ch.updatedAt
}

func (ch *htmlCreator) registerSub() (int, chan bool) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	if ch.subs == nil {
		ch.subs = make(map[int]chan bool)
	}

	ch.noOfSubs++
	id := ch.noOfSubs
	ch.subs[id] = make(chan bool)

	return id, ch.subs[id]
}

func (ch *htmlCreator) deRegisterSub(id int) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	delete(ch.subs, id)
}

const (
	contentType = "Content-Type"
	mimeHTML    = "text/html"
)

func handler(contentHolder *htmlCreator) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "HEAD" {
			handleHEAD(contentHolder, w, req)
			return
		}

		w.Header().Set(contentType, mimeHTML)

		content, _ := contentHolder.getHTMLContent()
		w.Write(content) // nolint: errcheck
	})

	return mux
}

func handleHEAD(contentHolder *htmlCreator, w http.ResponseWriter, req *http.Request) {
	content := []byte("")

	etag := req.Header.Get("If-Modified-Since")
	if etag == "" {
		w.Write(content) // nolint: errcheck
		return
	}

	id, contentChan := contentHolder.registerSub()
	defer contentHolder.deRegisterSub(id)

	_, updatedAt := contentHolder.getHTMLContent()

	since, err := time.Parse(time.RFC1123, etag)
	if err != nil {
		w.Write(content) // nolint: errcheck
		return
	}

	if since.Before(updatedAt) {
		w.Write(content) // nolint: errcheck
		return
	}

	const longPollingTimeout = 60 * time.Second

	select {
	case <-contentChan:
	case <-time.After(longPollingTimeout):
		w.WriteHeader(http.StatusNotModified)
	}

	w.Write(content) // nolint: errcheck
}

func buildHTML(files map[string]compiledFile) (_ []byte, err error) {
	generator, err := template.New("html_page").Parse(htmlPageTemplate)
	if err != nil {
		err = fmt.Errorf("failed to parse HTML page template: %w", err)
		return
	}

	var templateInfo = struct{ Files []compiledFile }{Files: make([]compiledFile, 0, len(files))}

	for _, file := range files {
		templateInfo.Files = append(templateInfo.Files, file)
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
		Static <a href="{{.PngLink}}">.png link</a> from planttext.com.
    <p>
      <img style="object-fit:contain;" src="{{.PngLink}}" alt=".png" />
    </p>

    <h3>.svg</h3>
		Static <a href="{{.SvgLink}}">.svg link</a> from planttext.com.
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
