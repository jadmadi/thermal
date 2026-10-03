# Facts: Platform-Native Application Data Directory Resolution

## Architectural Invariants & Constraints
- Existing Unix paths (`~/.local/share`, `~/.<tool>`) remain primary to preserve backward compatibility.
- Check Windows `%APPDATA%` and `%LOCALAPPDATA%` environment variables when running on Windows.
- Check macOS `~/Library/Application Support` directory when running on Darwin.
- No third-party dependencies: use Go standard library `os.UserConfigDir`, `os.UserHomeDir`, and `os.Getenv`.

## File & Interface Contracts
- Relevant files:
  - `internal/loaders/registry.go`: tool data paths and directory definitions.
  - `internal/thermal/time.go`: home directory and path utilities.

