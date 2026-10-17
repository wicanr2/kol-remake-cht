package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"
)

// A double-clicked Windows GUI process has no visible stderr. Keep its errors
// beside the user's saves; command-line diagnostics keep their original streams.
func startRuntimeLog(interactive, diagnostics bool, started time.Time) {
	if !interactive || diagnostics || runtime.GOOS != "windows" {
		return
	}
	savePath, err := defaultSavePath()
	if err != nil {
		return
	}
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}
	path := filepath.Join(dir, "kol.log")
	previous := filepath.Join(dir, "kol-previous.log")
	if _, err := os.Stat(path); err == nil {
		os.Remove(previous)
		os.Rename(path, previous)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	os.Stderr, os.Stdout = f, f
	debug.SetCrashOutput(f, debug.CrashOptions{})
	fmt.Fprintf(f, "武士傳說 %s %s %s/%s\n", buildVersion, started.Format(time.RFC3339), runtime.GOOS, runtime.GOARCH)
}
