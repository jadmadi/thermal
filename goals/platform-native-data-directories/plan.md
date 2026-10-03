# Plan: Platform-Native Application Data Directory Resolution

## Execution Steps
- [ ] Phase 1: Platform Path Resolver: Add platform-aware data directory resolution in `registry.go` probing OS standard directories.
- [ ] Phase 2: Tool Registry Integration: Wire discovery fallbacks for tools that support desktop installs on Windows and macOS.
- [ ] Phase 3: Unit Testing: Add unit tests with mock environment variables verifying platform directory fallback resolution.
- [ ] Phase 4: Attribution Commit: Commit with `(goals/platform-native-data-directories/goal.md)` using a `feat:` or `fix:` subject, then run `sila goals`.

