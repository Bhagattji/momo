matha — minimal scaffold

This is the initial scaffold for the matha CLI project. Start by running:

  go build -o matha.exe ./cmd

Files:
  cmd/main.go           — CLI entry
  internal/config       — basic config loader
  internal/version      — version constant

Next: implement provider registry, agent loop, and TUI.
