package task

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

type FileStore struct {
	path  string
	files []string
	rw    sync.RWMutex
}

type FileJanitor struct {
	store           *FileStore
	cleanupInterval time.Duration
	fileTTL         time.Duration
	cleanupTicker   *time.Ticker
}

func NewFileStore(path string) *FileStore {
	return &FileStore{
		path: path,
	}
}

func NewFileJanitor(store *FileStore, cleanupInterval, fileTTL time.Duration) *FileJanitor {
	return &FileJanitor{
		store:           store,
		cleanupInterval: cleanupInterval,
		fileTTL:         fileTTL,
	}
}

func (fs *FileStore) Create(filename string) (*os.File, error) {
	if err := os.MkdirAll(fs.path, 0755); err != nil {
		slog.Warn("failed to create upload directory", "dir", fs.path, "err", err)
		return nil, err
	}

	name := filepath.Join(fs.path, filepath.Base(filename))
	file, err := os.Create(name)

	if err != nil {
		slog.Warn("failed to create file", "file", name, "err", err)
		return nil, err
	}

	fs.rw.Lock()
	defer fs.rw.Unlock()
	fs.files = append(fs.files, name)

	return file, nil
}

func (fs *FileStore) Delete(name string) {
	if err := os.Remove(filepath.Join(fs.path, filepath.Base(name))); err != nil {
		slog.Warn("failed to remove file", "file", name, "err", err)
	}

	fs.untrack(name)
}

// untrack removes name from fs.files, taking fs.rw itself. Callers must not
// already hold fs.rw.
func (fs *FileStore) untrack(name string) {
	fs.rw.Lock()
	defer fs.rw.Unlock()
	if i := slices.Index(fs.files, name); i >= 0 {
		fs.files = slices.Delete(fs.files, i, i+1)
	}
}

func (fs *FileStore) Snapshot() []string {
	fs.rw.RLock()
	defer fs.rw.RUnlock()
	return slices.Clone(fs.files)
}

func (fj *FileJanitor) Start() {
	// preload dir contents
	entries, err := os.ReadDir(fj.store.path)
	if err != nil {
		slog.Warn("failed to read directory", "dir", fj.store.path, "err", err)
	} else {
		for _, file := range entries {
			if file.IsDir() {
				continue
			}

			fj.store.files = append(fj.store.files, filepath.Join(fj.store.path, filepath.Base(file.Name())))
		}
	}

	fj.cleanupTicker = time.NewTicker(fj.cleanupInterval)
	go func() {
		for range fj.cleanupTicker.C {
			fj.cleanup()
		}
	}()
}

func (fj *FileJanitor) Stop() {
	fj.cleanupTicker.Stop()
}

func (fj *FileJanitor) cleanup() {
	for _, name := range fj.store.Snapshot() {
		func() {
			file, err := os.Open(name)
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					slog.Warn("failed to close file", "file", name, "err", err)
				}
			}(file)

			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					fj.store.untrack(name)
				} else {
					slog.Error("failed to open file", "file", name, "err", err)
				}

				return
			}

			stat, err := file.Stat()
			if err != nil {
				slog.Warn("failed to get file stat", "file", name, "err", err)
				return
			}

			if time.Since(stat.ModTime()) > fj.fileTTL {
				slog.Info("file expired", "file", name)
				fj.store.Delete(name)
			}
		}()
	}
}
