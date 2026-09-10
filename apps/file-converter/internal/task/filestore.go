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
	path            string
	cleanupInterval time.Duration
	fileTTL         time.Duration
	files           []string
	cleanupTicker   *time.Ticker
	rw              sync.RWMutex
}

func NewFileStore(path string, cleanupInterval, fileTTL time.Duration) *FileStore {
	return &FileStore{
		path:            path,
		cleanupInterval: cleanupInterval,
		fileTTL:         fileTTL,
	}
}

func (fs *FileStore) Start() {
	// preload dir contents
	entries, err := os.ReadDir(fs.path)
	if err != nil {
		slog.Warn("failed to read directory", "dir", fs.path, "err", err)
	} else {
		for _, file := range entries {
			if file.IsDir() {
				continue
			}

			fs.files = append(fs.files, filepath.Join(fs.path, filepath.Base(file.Name())))
		}
	}

	fs.cleanupTicker = time.NewTicker(fs.cleanupInterval)
	go func() {
		for range fs.cleanupTicker.C {
			fs.cleanup()
		}
	}()
}

func (fs *FileStore) Stop() {
	fs.cleanupTicker.Stop()
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

func (fs *FileStore) cleanup() {
	fs.rw.RLock()
	names := slices.Clone(fs.files)
	fs.rw.RUnlock()

	for _, name := range names {
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
					fs.untrack(name)
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

			if time.Since(stat.ModTime()) > fs.fileTTL {
				slog.Info("file expired", "file", name)
				fs.Delete(name)
			}
		}()
	}
}
