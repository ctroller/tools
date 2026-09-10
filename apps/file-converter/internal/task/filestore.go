package task

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type FileStore struct {
	path string
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

	return file, nil
}

func (fs *FileStore) Delete(name string) {
	if err := os.Remove(filepath.Join(fs.path, filepath.Base(name))); err != nil {
		slog.Warn("failed to remove file", "file", name, "err", err)
	}
}

func (fj *FileJanitor) Start() {
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
	entries, err := os.ReadDir(fj.store.path)
	if err != nil {
		slog.Warn("failed to read directory", "dir", fj.store.path, "err", err)
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			func() {
				stat, err := entry.Info()
				if err != nil {
					slog.Warn("failed to get file stat", "file", entry.Name(), "err", err)
					return
				}

				if time.Since(stat.ModTime()) > fj.fileTTL {
					slog.Info("file expired", "file", entry.Name())
					fj.store.Delete(entry.Name())
				}
			}()
		}
	}
}
