package generator

import (
	"bytes"
	"path/filepath"
	"sync"
	"time"

	"github.com/lonnblad/gopuml"
)

type File struct {
	Filepath  string
	Filename  string
	UpdatedAt time.Time
	Raw       []byte
	Encoded   []byte
}

type Generator struct {
	files map[string]File
	subs  map[int]chan File

	noOfSubs int

	mutex sync.RWMutex
}

func New() *Generator {
	return &Generator{
		files: make(map[string]File),
		subs:  make(map[int]chan File),
	}
}

func (gen *Generator) PutFile(path string, rawContent []byte) error {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	oldFile := gen.files[path]
	if oldFile.Filepath == path && bytes.Equal(oldFile.Raw, rawContent) {
		return nil
	}

	compressed, err := gopuml.Deflate(rawContent)
	if err != nil {
		return err
	}

	encoded := gopuml.Encode(compressed)

	f := File{
		Filepath:  path,
		Filename:  filepath.Base(path),
		UpdatedAt: time.Now(),
		Raw:       rawContent,
		Encoded:   encoded,
	}

	gen.files[path] = f

	for _, sub := range gen.subs {
		sub <- f
	}

	return nil
}

func (gen *Generator) RegisterSub() (int, chan File) {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	gen.noOfSubs++
	id := gen.noOfSubs
	gen.subs[id] = make(chan File)

	return id, gen.subs[id]
}

func (gen *Generator) DeregisterSub(id int) {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	delete(gen.subs, id)
}

func (gen *Generator) GetFiles() []File {
	gen.mutex.RLock()
	defer gen.mutex.RUnlock()

	result := make([]File, 0, len(gen.files))

	for _, file := range gen.files {
		result = append(result, file)
	}

	return result
}
