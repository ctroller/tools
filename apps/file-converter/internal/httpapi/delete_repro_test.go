package httpapi

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"trox.dev/file-converter/internal/config"
	"trox.dev/file-converter/internal/convert"
	"trox.dev/file-converter/internal/task"
)

// DeleteHandle must delete the job's actual on-disk file, which is named after the job ID
// only once it has been converted; an uploaded-but-not-yet-converted job's file is named
// "<id>.tmp" and lives at result.FilePath.
func TestDeleteHandle_UploadedJob_RemovesFile(t *testing.T) {
	dir := t.TempDir()
	fs := task.NewFileStore(dir)
	store := task.NewStatusStore()
	registry := convert.NewRegistry()
	queue := task.NewQueue(1, 1, task.NewJobExecutor(fs), store)
	intake := task.NewJobIntake(registry, queue, store, fs)

	res, err := intake.Submit(strings.NewReader("hello"), "text/plain")
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file on disk after submit, got %d", len(entries))
	}
	onDiskBefore := filepath.Join(dir, entries[0].Name())

	api := &API{config: &config.Config{}, registry: registry, intake: intake, fs: fs}

	req := httptest.NewRequest("DELETE", "/files/"+res.JobID, nil)
	req.SetPathValue("handle", res.JobID)
	w := httptest.NewRecorder()
	api.DeleteHandle(w, req)

	if _, statErr := os.Stat(onDiskBefore); statErr == nil {
		t.Fatalf("file %s still exists after DeleteHandle (status %d)", onDiskBefore, w.Code)
	}
}

// A failed job must still carry a usable FilePath: DeleteHandle relies on it, and an
// empty FilePath resolves (via filepath.Base("")==".") to the store directory itself.
func TestDeleteHandle_FailedJob_RemovesFileWithoutTouchingStoreDir(t *testing.T) {
	dir := t.TempDir()
	fs := task.NewFileStore(dir)
	store := task.NewStatusStore()

	uploaded, err := fs.Create("input.tmp")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	uploaded.Close()

	store.Set("job-1", task.JobResult{JobID: "job-1", Status: task.StatusFailed, FilePath: uploaded.Name()})

	api := &API{config: &config.Config{}, intake: task.NewJobIntake(convert.NewRegistry(), task.NewQueue(1, 1, task.NewJobExecutor(fs), store), store, fs), fs: fs}

	req := httptest.NewRequest("DELETE", "/files/job-1", nil)
	req.SetPathValue("handle", "job-1")
	w := httptest.NewRecorder()
	api.DeleteHandle(w, req)

	if w.Code != 202 {
		t.Fatalf("expected 202, got %d", w.Code)
	}
	if _, statErr := os.Stat(uploaded.Name()); statErr == nil {
		t.Fatalf("file %s still exists after DeleteHandle", uploaded.Name())
	}
	if _, statErr := os.Stat(dir); statErr != nil {
		t.Fatalf("store directory %s was removed: %v", dir, statErr)
	}
}

// DeleteHandle must remove the job's record, not just its file: a second delete of the
// same handle must 404, and a stale handle must not be startable via Convert afterward.
func TestDeleteHandle_RemovesJobRecord(t *testing.T) {
	dir := t.TempDir()
	fs := task.NewFileStore(dir)
	store := task.NewStatusStore()
	registry := convert.NewRegistry()
	queue := task.NewQueue(1, 1, task.NewJobExecutor(fs), store)
	intake := task.NewJobIntake(registry, queue, store, fs)

	res, err := intake.Submit(strings.NewReader("hello"), "text/plain")
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	api := &API{config: &config.Config{}, registry: registry, intake: intake, fs: fs}

	req := httptest.NewRequest("DELETE", "/files/"+res.JobID, nil)
	req.SetPathValue("handle", res.JobID)
	api.DeleteHandle(httptest.NewRecorder(), req)

	if _, found := intake.Lookup(res.JobID); found {
		t.Fatalf("job record for %s still present after delete", res.JobID)
	}

	w := httptest.NewRecorder()
	api.DeleteHandle(w, req)
	if w.Code != 404 {
		t.Fatalf("expected 404 on second delete, got %d", w.Code)
	}
}
