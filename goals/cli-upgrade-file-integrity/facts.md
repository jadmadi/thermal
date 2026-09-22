# Facts: Archive Extraction Deduplication and File Write Close Verification in CLI Self-Updater

## Architectural Facts & Grounding

1. **File Close on Write**:
   In Go, calling `defer file.Close()` is sufficient for read-only files, but dangerous for files opened for writing (`os.Create`, `os.OpenFile`). Operating systems buffer write operations; any error encountered when flushing dirty buffers or committing directory metadata (such as `ENOSPC` disk full or remote I/O failure) is returned by `Close()`. Therefore, writable files must inspect the error from `Close()` before reporting success.

2. **Deduplication in Tar Reader**:
   In `extractBinary`, `filepath.Base(hdr.Name) == "thermal"` uniquely identifies the executable inside the GoReleaser `.tar.gz` archive regardless of whether it resides at archive root or in an archive folder. Redundant conditions checking `!strings.Contains(hdr.Name, "/")` duplicated 10 lines of code without altering behavior.
